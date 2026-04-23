package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	awspricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

// The Pricing API endpoint is only served from us-east-1 and ap-south-1, but it returns pricing
// data for every AWS region via the regionCode filter on each query. We always talk to us-east-1
// since it is available to every AWS account.
const pricingEndpointRegion = "us-east-1"

// maxPricingConcurrency caps parallel GetProducts calls during Load.
const maxPricingConcurrency = 8

// hoursPerMonth converts Pricing API hourly rates into an approximate monthly cost. 730 is the
// same figure AWS uses in Cost Explorer examples (365.25 * 24 / 12).
const hoursPerMonth = 730.0

// Cache keys are flat strings, usually "category" or "category:variant" (e.g. "nat",
// "ebs:gp3", "rds-instance:db.t3.medium").
const (
	categoryEBS         = "ebs"
	categoryEIP         = "eip"
	categoryNAT         = "nat"
	categoryLBApp       = "lb-app"
	categoryLBClassic   = "lb-classic"
	categoryCWLogs      = "cwlogs"
	categoryRDSInstance = "rds-instance"
	categoryRDSStorage  = "rds-storage"
	categoryRDSSnapshot = "rds-snapshot"
)

const (
	// EIPCostPerMonth is the default cost of an unassociated Elastic IP (~$0.005/hour * 730 hours).
	EIPCostPerMonth = 3.65

	// EBSgp2CostPerGBMonth is the default cost of gp2 EBS storage per GB per month.
	EBSgp2CostPerGBMonth = 0.10
	// EBSgp3CostPerGBMonth is the default cost of gp3 EBS storage per GB per month.
	EBSgp3CostPerGBMonth = 0.08
	// EBSio1CostPerGBMonth is the default cost of io1 EBS storage per GB per month.
	EBSio1CostPerGBMonth = 0.125
	// EBSio2CostPerGBMonth is the default cost of io2 EBS storage per GB per month.
	EBSio2CostPerGBMonth = 0.125
	// EBSst1CostPerGBMonth is the default cost of st1 EBS storage per GB per month.
	EBSst1CostPerGBMonth = 0.045
	// EBSsc1CostPerGBMonth is the default cost of sc1 EBS storage per GB per month.
	EBSsc1CostPerGBMonth = 0.015

	// CloudWatchLogsCostPerGBMonth is the default cost of CloudWatch Logs storage per GB per month.
	CloudWatchLogsCostPerGBMonth = 0.03

	// NatGatewayCostPerMonth is the default cost of a NAT Gateway (~$0.045/hour * 730 hours in us-east-1).
	NatGatewayCostPerMonth = 32.85

	// ALBCostPerMonth is the default base cost of an ALB/NLB (~$0.0225/hour * 730 hours).
	ALBCostPerMonth = 16.43

	// CLBCostPerMonth is the default base cost of a CLB (~$0.025/hour * 730 hours).
	CLBCostPerMonth = 18.25

	// RDSStorageCostPerGBMonth is the default cost of RDS gp2 storage per GB per month.
	RDSStorageCostPerGBMonth = 0.115
	// RDSSnapshotCostPerGBMonth is the default cost of RDS snapshot storage per GB per month.
	RDSSnapshotCostPerGBMonth = 0.095

	// EBSSnapshotCostPerGBMonth is the default cost of EBS snapshot storage per GB per month.
	EBSSnapshotCostPerGBMonth = 0.05
)

// ebsSpec maps each supported EBS volume type to its Pricing API variant name and default rate.
var ebsSpec = map[types.VolumeType]struct {
	variant     string
	defaultRate float64
}{
	types.VolumeTypeGp2: {"gp2", EBSgp2CostPerGBMonth},
	types.VolumeTypeGp3: {"gp3", EBSgp3CostPerGBMonth},
	types.VolumeTypeIo1: {"io1", EBSio1CostPerGBMonth},
	types.VolumeTypeIo2: {"io2", EBSio2CostPerGBMonth},
	types.VolumeTypeSt1: {"st1", EBSst1CostPerGBMonth},
	types.VolumeTypeSc1: {"sc1", EBSsc1CostPerGBMonth},
}

// rdsInstancePricing maps RDS instance classes to approximate monthly compute costs (us-east-1,
// Single-AZ, On-Demand MySQL).
var rdsInstancePricing = map[string]float64{
	"db.t3.micro": 12.41, "db.t3.small": 24.82, "db.t3.medium": 49.64, "db.t3.large": 99.28, "db.t3.xlarge": 198.56,
	"db.t4g.micro": 11.52, "db.t4g.small": 23.04, "db.t4g.medium": 46.08, "db.t4g.large": 92.16, "db.t4g.xlarge": 184.32,
	"db.m5.large": 124.10, "db.m5.xlarge": 248.20, "db.m5.2xlarge": 496.40, "db.m5.4xlarge": 992.80, "db.m5.8xlarge": 1985.60,
	"db.m6g.large": 118.26, "db.m6g.xlarge": 236.52, "db.m6g.2xlarge": 473.04, "db.m6g.4xlarge": 946.08, "db.m6g.8xlarge": 1892.16,
	"db.m7g.large": 127.02, "db.m7g.xlarge": 254.04, "db.m7g.2xlarge": 508.08, "db.m7g.4xlarge": 1016.16, "db.m7g.8xlarge": 2032.32,
	"db.r5.large": 175.20, "db.r5.xlarge": 350.40, "db.r5.2xlarge": 700.80, "db.r5.4xlarge": 1401.60, "db.r5.8xlarge": 2803.20,
	"db.r6g.large": 166.44, "db.r6g.xlarge": 332.88, "db.r6g.2xlarge": 665.76, "db.r6g.4xlarge": 1331.52, "db.r6g.8xlarge": 2663.04,
	"db.r7g.large": 178.85, "db.r7g.xlarge": 357.70, "db.r7g.2xlarge": 715.40, "db.r7g.4xlarge": 1430.80, "db.r7g.8xlarge": 2861.60,
}

var sagemakerInstancePricing = map[string]float64{
	"ml.t2.medium": 46.72, "ml.t2.large": 93.44, "ml.t2.xlarge": 186.15, "ml.t2.2xlarge": 372.30,
	"ml.t3.medium": 41.61, "ml.t3.large": 83.22, "ml.t3.xlarge": 166.44, "ml.t3.2xlarge": 332.88,
	"ml.m4.xlarge": 204.40, "ml.m4.2xlarge": 408.80, "ml.m4.4xlarge": 817.60,
	"ml.m5.large": 101.91, "ml.m5.xlarge": 203.82, "ml.m5.2xlarge": 407.64, "ml.m5.4xlarge": 815.28,
	"ml.c4.xlarge": 176.66, "ml.c4.2xlarge": 353.32, "ml.c4.4xlarge": 706.64,
	"ml.c5.large": 87.60, "ml.c5.xlarge": 175.20, "ml.c5.2xlarge": 350.40, "ml.c5.4xlarge": 700.80,
	"ml.p2.xlarge": 1144.80, "ml.g4dn.xlarge": 538.72, "ml.g5.xlarge": 802.82,
	"ml.inf1.xlarge": 218.40,
}

// NewService creates a new pricing service.
func NewService(awsconfig aws.Config) Service {
	cfg := awsconfig.Copy()
	region := awsconfig.Region
	cfg.Region = pricingEndpointRegion
	client := awspricing.NewFromConfig(cfg)

	return &service{
		client: client,
		prices: make(map[string]float64),
		region: region,
	}
}

func (s *service) LoadRegionRates(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxPricingConcurrency)

	var (
		errMu    sync.Mutex
		fetchErr []error
	)

	fetch := func(category, serviceCode string, filters []pricingtypes.Filter, extract func(attrs map[string]string) (variant string, ok bool)) {
		g.Go(func() error {
			entries, err := s.queryProducts(ctx, serviceCode, filters)
			if err != nil {
				wrapped := fmt.Errorf("%s: %w", category, err)
				errMu.Lock()
				fetchErr = append(fetchErr, wrapped)
				errMu.Unlock()
				return nil
			}

			for _, entry := range entries {
				variant, ok := extract(entry.attrs)
				if !ok {
					continue
				}
				s.setPrice(priceKey(category, variant), entry.pricePerUnit)
			}
			return nil
		})
	}

	regionFilter := pricingtypes.Filter{
		Type:  pricingtypes.FilterTypeTermMatch,
		Field: aws.String("regionCode"),
		Value: aws.String(s.region),
	}

	matchUsagetypeContains := func(needle string) func(attrs map[string]string) (string, bool) {
		return func(attrs map[string]string) (string, bool) {
			if !strings.Contains(attrs["usagetype"], needle) {
				return "", false
			}
			return "", true
		}
	}

	matchLBUsage := func(attrs map[string]string) (string, bool) {
		u := attrs["usagetype"]
		if !strings.Contains(u, "LoadBalancerUsage") || strings.Contains(u, "Outposts-") || strings.Contains(u, "TS-") {
			return "", false
		}
		return "", true
	}

	for _, v := range []string{"gp2", "gp3", "io1", "io2", "st1", "sc1"} {
		v := v
		fetch(categoryEBS, "AmazonEC2", []pricingtypes.Filter{
			regionFilter,
			termMatch("productFamily", "Storage"),
			termMatch("volumeApiName", v),
		}, func(attrs map[string]string) (string, bool) {
			return attrs["volumeApiName"], attrs["volumeApiName"] != ""
		})
	}

	fetch(categoryEIP, "AmazonVPC", []pricingtypes.Filter{
		regionFilter,
		termMatch("group", "VPCPublicIPv4Address"),
	}, matchUsagetypeContains("PublicIPv4:IdleAddress"))

	fetch(categoryNAT, "AmazonEC2", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "NAT Gateway"),
		termMatch("operation", "NatGateway"),
	}, matchUsagetypeContains("NatGateway-Hours"))

	lbFamilies := map[string]string{
		"Load Balancer-Application": "application",
		"Load Balancer-Network":     "network",
	}

	for family, variant := range lbFamilies {
		family, variant := family, variant
		fetch(categoryLBApp, "AWSELB", []pricingtypes.Filter{
			regionFilter,
			termMatch("productFamily", family),
		}, func(attrs map[string]string) (string, bool) {
			if _, ok := matchLBUsage(attrs); !ok {
				return "", false
			}
			return variant, true
		})
	}

	fetch(categoryLBClassic, "AWSELB", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Load Balancer"),
	}, matchLBUsage)

	fetch(categoryCWLogs, "AmazonCloudWatch", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Storage Snapshot"),
	}, matchUsagetypeContains("TimedStorage-ByteHrs"))

	fetch(categoryRDSStorage, "AmazonRDS", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Database Storage"),
		termMatch("volumeType", "General Purpose"),
		termMatch("deploymentOption", "Single-AZ"),
	}, matchUsagetypeContains("RDS:GP2-Storage"))

	fetch(categoryRDSSnapshot, "AmazonRDS", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Storage Snapshot"),
	}, matchUsagetypeContains("RDS:ChargedBackupUsage"))

	fetch(categoryRDSInstance, "AmazonRDS", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Database Instance"),
		termMatch("databaseEngine", "MySQL"),
		termMatch("deploymentOption", "Single-AZ"),
	}, func(attrs map[string]string) (string, bool) {
		return attrs["instanceType"], attrs["instanceType"] != ""
	})

	_ = g.Wait()

	return errors.Join(fetchErr...)
}

func (s *service) CalculateEBSMonthlyCost(sizeGiB int32, volumeType types.VolumeType) float64 {
	spec, ok := ebsSpec[volumeType]
	if !ok {
		spec = ebsSpec[types.VolumeTypeGp2]
	}

	if v, ok := s.lookupMonthly(priceKey(categoryEBS, spec.variant), 0); ok {
		return float64(sizeGiB) * v
	}

	return float64(sizeGiB) * spec.defaultRate
}

func (s *service) CalculateEBSSnapshotMonthlyCost(sizeGB int64) float64 {
	return float64(sizeGB) * EBSSnapshotCostPerGBMonth
}

func (s *service) CalculateEIPMonthlyCost() float64 {
	if v, ok := s.lookupMonthly(priceKey(categoryEIP, ""), hoursPerMonth); ok {
		return v
	}
	return EIPCostPerMonth
}

func (s *service) CalculateLoadBalancerMonthlyCost(lbType elbtypes.LoadBalancerTypeEnum) float64 {
	if lbType == "classic" {
		if v, ok := s.lookupMonthly(priceKey(categoryLBClassic, ""), hoursPerMonth); ok {
			return v
		}
		return CLBCostPerMonth
	}

	if v, ok := s.lookupMonthly(priceKey(categoryLBApp, string(lbType)), hoursPerMonth); ok {
		return v
	}
	return ALBCostPerMonth
}

func (s *service) CalculateCloudWatchLogsMonthlyCost(storedBytes int64) float64 {
	storedGB := float64(storedBytes) / (1024 * 1024 * 1024)
	rate := CloudWatchLogsCostPerGBMonth
	if v, ok := s.lookupMonthly(priceKey(categoryCWLogs, ""), 0); ok {
		rate = v
	}
	return storedGB * rate
}

func (s *service) CalculateNATGatewayMonthlyCost() float64 {
	if v, ok := s.lookupMonthly(priceKey(categoryNAT, ""), hoursPerMonth); ok {
		return v
	}
	return NatGatewayCostPerMonth
}

func (s *service) CalculateRDSInstanceMonthlyCost(allocatedGB int32, multiAZ bool) float64 {
	cost := float64(allocatedGB) * s.rdsStoragePerGBMonth()
	if multiAZ {
		cost *= 2
	}
	return cost
}

func (s *service) CalculateRDSSnapshotMonthlyCost(allocatedGB int32) float64 {
	return float64(allocatedGB) * s.rdsSnapshotPerGBMonth()
}

func (s *service) CalculateRDSIdleInstanceMonthlyCost(instanceClass string, allocatedGB int32, multiAZ bool) float64 {
	computeCost := s.rdsInstanceComputeCost(instanceClass)
	storageCost := float64(allocatedGB) * s.rdsStoragePerGBMonth()
	total := computeCost + storageCost
	if multiAZ {
		total *= 2
	}
	return total
}

func (s *service) CalculateSageMakerEndpointMonthlyCost(variants []model.SageMakerVariant) float64 {
	var total float64
	for _, v := range variants {
		total += sagemakerInstancePricing[v.InstanceType] * float64(v.InstanceCount)
	}
	return total
}

// Internal helpers

func (s *service) rdsStoragePerGBMonth() float64 {
	if v, ok := s.lookupMonthly(priceKey(categoryRDSStorage, ""), 0); ok {
		return v
	}
	return RDSStorageCostPerGBMonth
}

func (s *service) rdsSnapshotPerGBMonth() float64 {
	if v, ok := s.lookupMonthly(priceKey(categoryRDSSnapshot, ""), 0); ok {
		return v
	}
	return RDSSnapshotCostPerGBMonth
}

func (s *service) rdsInstanceComputeCost(instanceClass string) float64 {
	if v, ok := s.lookupMonthly(priceKey(categoryRDSInstance, instanceClass), hoursPerMonth); ok {
		return v
	}
	if trimmed := strings.TrimPrefix(instanceClass, "db."); trimmed != instanceClass {
		if v, ok := s.lookupMonthly(priceKey(categoryRDSInstance, trimmed), hoursPerMonth); ok {
			return v
		}
	}
	return rdsInstancePricing[instanceClass]
}

func (s *service) setPrice(k string, v float64) {
	s.priceMu.Lock()
	defer s.priceMu.Unlock()
	s.prices[k] = v
}

func (s *service) lookupMonthly(k string, hoursInMonth float64) (float64, bool) {
	s.priceMu.RLock()
	defer s.priceMu.RUnlock()
	v, ok := s.prices[k]
	if !ok {
		return 0, false
	}
	if hoursInMonth > 0 {
		return v * hoursInMonth, true
	}
	return v, true
}

type productEntry struct {
	attrs        map[string]string
	pricePerUnit float64
}

func (s *service) queryProducts(ctx context.Context, serviceCode string, filters []pricingtypes.Filter) ([]productEntry, error) {
	input := &awspricing.GetProductsInput{
		ServiceCode:   aws.String(serviceCode),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
	}
	paginator := awspricing.NewGetProductsPaginator(s.client, input)
	var out []productEntry
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("GetProducts %s: %w", serviceCode, err)
		}
		for _, raw := range resp.PriceList {
			if entry, ok := parsePriceListDocument(raw); ok {
				out = append(out, entry)
			}
		}
	}
	return out, nil
}

func parsePriceListDocument(raw string) (productEntry, bool) {
	var doc struct {
		Product struct {
			Attributes map[string]string `json:"attributes"`
		} `json:"product"`
		Terms struct {
			OnDemand map[string]struct {
				PriceDimensions map[string]struct {
					PricePerUnit map[string]string `json:"pricePerUnit"`
				} `json:"priceDimensions"`
			} `json:"OnDemand"`
		} `json:"terms"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return productEntry{}, false
	}
	for _, sku := range doc.Terms.OnDemand {
		for _, dim := range sku.PriceDimensions {
			usdStr, ok := dim.PricePerUnit["USD"]
			if !ok {
				continue
			}
			usd, err := strconv.ParseFloat(usdStr, 64)
			if err != nil {
				continue
			}
			return productEntry{attrs: doc.Product.Attributes, pricePerUnit: usd}, true
		}
	}
	return productEntry{}, false
}

func priceKey(category, variant string) string {
	if variant == "" {
		return category
	}
	return category + ":" + variant
}

func termMatch(field, value string) pricingtypes.Filter {
	return pricingtypes.Filter{
		Type:  pricingtypes.FilterTypeTermMatch,
		Field: aws.String(field),
		Value: aws.String(value),
	}
}
