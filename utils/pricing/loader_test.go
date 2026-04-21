package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/stretchr/testify/assert"
)

type fakePricingClient struct {
	products map[string][]map[string]any
	errOn    map[string]error
}

func (f *fakePricingClient) GetProducts(_ context.Context, in *pricing.GetProductsInput, _ ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	svc := aws.ToString(in.ServiceCode)
	key := svc + ":" + filterKey(in.Filters)

	if err := f.errOn[svc]; err != nil {
		return nil, err
	}

	list := f.products[key]
	if list == nil {
		list = f.products[svc]
	}

	out := make([]string, 0, len(list))

	for _, doc := range list {
		b, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}

		out = append(out, string(b))
	}

	return &pricing.GetProductsOutput{PriceList: out}, nil
}

func filterKey(filters []pricingtypes.Filter) string {
	parts := make([]string, 0, len(filters))
	for _, f := range filters {
		parts = append(parts, aws.ToString(f.Field)+"="+aws.ToString(f.Value))
	}

	return strings.Join(parts, ";")
}

// buildProductDoc builds a minimal PriceList document shaped like the Pricing API JSON.
func buildProductDoc(attrs map[string]string, usdPerUnit string) map[string]any {
	return map[string]any{
		"product": map[string]any{
			"attributes": attrs,
		},
		"terms": map[string]any{
			"OnDemand": map[string]any{
				"sku.abc": map[string]any{
					"priceDimensions": map[string]any{
						"sku.abc.dim": map[string]any{
							"pricePerUnit": map[string]string{
								"USD": usdPerUnit,
							},
						},
					},
				},
			},
		},
	}
}

func TestLoadWithClient_PopulatesAllCategories(t *testing.T) {
	resetForTest()
	defer resetForTest()

	client := &fakePricingClient{products: map[string][]map[string]any{}}

	// EBS entries, one per volume type.
	for _, v := range []string{"gp2", "gp3", "io1", "io2", "st1", "sc1"} {
		key := "AmazonEC2:regionCode=eu-west-1;productFamily=Storage;volumeApiName=" + v
		client.products[key] = []map[string]any{
			buildProductDoc(map[string]string{"volumeApiName": v}, "0.08"),
		}
	}

	// EIP: AmazonVPC with usagetype containing "PublicIPv4:IdleAddress" (not the InUse variant).
	client.products["AmazonVPC:regionCode=eu-west-1;group=VPCPublicIPv4Address"] = []map[string]any{
		buildProductDoc(map[string]string{"usagetype": "EU-PublicIPv4:InUseAddress"}, "0.005"),
		buildProductDoc(map[string]string{"usagetype": "EU-PublicIPv4:IdleAddress"}, "0.006"),
	}

	// NAT: usagetype contains "NatGateway-Hours" (skip Bytes/Prvd rows).
	client.products["AmazonEC2:regionCode=eu-west-1;productFamily=NAT Gateway;operation=NatGateway"] = []map[string]any{
		buildProductDoc(map[string]string{"usagetype": "EU-NatGateway-Bytes"}, "0.045"),
		buildProductDoc(map[string]string{"usagetype": "EU-NatGateway-Hours"}, "0.05"),
	}

	// ALB: matchLBUsage accepts LoadBalancerUsage rows, rejects LCU/TS/Outposts.
	client.products["AWSELB:regionCode=eu-west-1;productFamily=Load Balancer-Application"] = []map[string]any{
		buildProductDoc(map[string]string{"usagetype": "EU-LCUUsage"}, "0.008"),
		buildProductDoc(map[string]string{"usagetype": "EU-LoadBalancerUsage"}, "0.025"),
		buildProductDoc(map[string]string{"usagetype": "EU-Outposts-LoadBalancerUsage"}, "0.025"),
		buildProductDoc(map[string]string{"usagetype": "EU-TS-LoadBalancerUsage"}, "0.005"),
	}

	// CLB: same matchLBUsage logic.
	client.products["AWSELB:regionCode=eu-west-1;productFamily=Load Balancer"] = []map[string]any{
		buildProductDoc(map[string]string{"usagetype": "EU-LoadBalancerUsage"}, "0.028"),
	}

	// CW Logs: usagetype contains "TimedStorage-ByteHrs".
	client.products["AmazonCloudWatch:regionCode=eu-west-1;productFamily=Storage Snapshot"] = []map[string]any{
		buildProductDoc(map[string]string{"usagetype": "EU-TimedStorage-ByteHrs"}, "0.033"),
	}

	// RDS storage: usagetype contains "RDS:GP2-Storage" (matches both prefixed and non).
	client.products["AmazonRDS:regionCode=eu-west-1;productFamily=Database Storage;volumeType=General Purpose;deploymentOption=Single-AZ"] = []map[string]any{
		buildProductDoc(map[string]string{"usagetype": "EU-RDS:GP2-Storage"}, "0.13"),
	}

	// RDS snapshot: usagetype contains "RDS:ChargedBackupUsage", NOT "RDSCustom:...".
	client.products["AmazonRDS:regionCode=eu-west-1;productFamily=Storage Snapshot"] = []map[string]any{
		buildProductDoc(map[string]string{"usagetype": "EU-RDSCustom:ChargedBackupUsage"}, "0.20"),
		buildProductDoc(map[string]string{"usagetype": "EU-RDS:ChargedBackupUsage"}, "0.10"),
	}

	// RDS instance (MySQL): keyed by instanceType, no usagetype filter.
	client.products["AmazonRDS:regionCode=eu-west-1;productFamily=Database Instance;databaseEngine=MySQL;deploymentOption=Single-AZ"] = []map[string]any{
		buildProductDoc(map[string]string{"instanceType": "db.t3.medium"}, "0.068"),
	}

	loadWithClient(context.Background(), client, "eu-west-1")

	assert.InDelta(t, 0.08, EBSCostPerGBMonth(types.VolumeTypeGp3), 1e-9)
	assert.InDelta(t, 0.006*hoursPerMonth, CalculateEIPMonthlyCost(), 1e-6)
	assert.InDelta(t, 0.05*hoursPerMonth, CalculateNATGatewayMonthlyCost(), 1e-6)
	assert.InDelta(t, 0.025*hoursPerMonth, CalculateLoadBalancerMonthlyCost("application"), 1e-6)
	assert.InDelta(t, 0.028*hoursPerMonth, CalculateLoadBalancerMonthlyCost("classic"), 1e-6)

	// 1 GiB of CW Logs at 0.033/GB-mo
	assert.InDelta(t, 0.033, CalculateCloudWatchLogsMonthlyCost(1024*1024*1024), 1e-6)

	// RDS instance: 0.068/hr * 730 + 10 GB * 0.13
	assert.InDelta(t, 0.068*hoursPerMonth+10*0.13, CalculateRDSIdleInstanceMonthlyCost("db.t3.medium", 10, false), 1e-6)
}

func TestLoadWithClient_FallsBackWhenClientErrors(t *testing.T) {
	resetForTest()
	defer resetForTest()

	client := &fakePricingClient{
		products: map[string][]map[string]any{},
		errOn: map[string]error{
			"AmazonEC2":        errors.New("throttled"),
			"AWSELB":           errors.New("throttled"),
			"AmazonCloudWatch": errors.New("throttled"),
			"AmazonRDS":        errors.New("throttled"),
		},
	}

	loadWithClient(context.Background(), client, "us-west-2")

	assert.InDelta(t, EBSgp3CostPerGBMonth, EBSCostPerGBMonth(types.VolumeTypeGp3), 1e-9)
	assert.InDelta(t, EIPCostPerMonth, CalculateEIPMonthlyCost(), 1e-9)
	assert.InDelta(t, NatGatewayCostPerMonth, CalculateNATGatewayMonthlyCost(), 1e-9)
	assert.InDelta(t, CloudWatchLogsCostPerGBMonth, CalculateCloudWatchLogsMonthlyCost(1024*1024*1024), 1e-9)
}

func TestLoadWithClient_PartialFailure(t *testing.T) {
	resetForTest()
	defer resetForTest()

	client := &fakePricingClient{
		products: map[string][]map[string]any{
			"AmazonEC2:regionCode=us-east-1;productFamily=NAT Gateway;operation=NatGateway": {
				buildProductDoc(map[string]string{"usagetype": "USE1-NatGateway-Hours"}, "0.045"),
			},
		},
		errOn: map[string]error{
			"AWSELB": errors.New("throttled"),
		},
	}

	loadWithClient(context.Background(), client, "us-east-1")

	assert.InDelta(t, 0.045*hoursPerMonth, CalculateNATGatewayMonthlyCost(), 1e-6)
	assert.InDelta(t, ALBCostPerMonth, CalculateLoadBalancerMonthlyCost("application"), 1e-9)
}

func TestParsePriceListDocument(t *testing.T) {
	doc := buildProductDoc(map[string]string{"instanceType": "db.m5.large"}, "0.17")

	b, _ := json.Marshal(doc)
	entry, ok := parsePriceListDocument(string(b))

	assert.True(t, ok)
	assert.Equal(t, "db.m5.large", entry.attrs["instanceType"])
	assert.InDelta(t, 0.17, entry.pricePerUnit, 1e-9)
}

func TestParsePriceListDocument_InvalidJSON(t *testing.T) {
	_, ok := parsePriceListDocument("{not json")
	assert.False(t, ok)
}

func TestParsePriceListDocument_UnparseableUSD(t *testing.T) {
	doc := buildProductDoc(map[string]string{}, "not-a-number")
	b, _ := json.Marshal(doc)
	_, ok := parsePriceListDocument(string(b))
	assert.False(t, ok)
}

func TestQueryProductsPagination(t *testing.T) {
	resetForTest()
	defer resetForTest()

	client := &paginatingClient{
		pages: []*pricing.GetProductsOutput{
			{PriceList: []string{mustJSON(t, buildProductDoc(map[string]string{"volumeApiName": "gp2"}, "0.10"))}, NextToken: aws.String("tok")},
			{PriceList: []string{mustJSON(t, buildProductDoc(map[string]string{"volumeApiName": "gp2"}, "0.11"))}, NextToken: nil},
		},
	}

	entries, err := queryProducts(context.Background(), client, "AmazonEC2", nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 2)
}

type paginatingClient struct {
	pages []*pricing.GetProductsOutput
	idx   int
}

func (p *paginatingClient) GetProducts(_ context.Context, _ *pricing.GetProductsInput, _ ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	if p.idx >= len(p.pages) {
		return nil, fmt.Errorf("unexpected call")
	}

	out := p.pages[p.idx]
	p.idx++

	return out, nil
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	return string(b)
}
