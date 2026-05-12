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

	//nolint:unparam
	matchUsagetypeContains := func(needle string) func(attrs map[string]string) (string, bool) {
		return func(attrs map[string]string) (string, bool) {
			ut := attrs["usagetype"]
			if !strings.Contains(ut, needle) {
				return "", false
			}

			return "", true
		}
	}

	//nolint:unparam
	matchLBUsage := func(attrs map[string]string) (string, bool) {
		ut := attrs["usagetype"]
		if !strings.Contains(ut, "LoadBalancerUsage") || strings.Contains(ut, "Outposts-") || strings.Contains(ut, "TS-") {
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

	// SageMaker real-time inference (component=Hosting). The Pricing API exposes the instance
	// type under the `instanceName` attribute (e.g. "ml.m5.xlarge"), which matches the value
	// returned by DescribeEndpointConfig used by the orchestrator.
	fetch(categorySageMakerHosting, "AmazonSageMaker", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "ML Instance"),
		termMatch("component", "Hosting"),
	}, func(attrs map[string]string) (string, bool) {
		return attrs["instanceName"], attrs["instanceName"] != ""
	})

	fetch(categoryECR, "AmazonECR", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "EC2 Container Registry"),
	}, matchUsagetypeContains("TimedStorage-ByteHrs"))

	// EC2 on-demand compute instances. The Pricing API exposes the instance type under the
	// `instanceType` attribute. Filter to Linux/Shared tenancy with no preinstalled software so
	// the cached rate represents a plain on-demand baseline. `capacitystatus=Used` is load
	// bearing: without it the API returns extra SKUs for reserved capacity reservations and
	// `parsePriceListDocument` would non-deterministically pick whichever comes first.
	fetch(categoryEC2Instance, "AmazonEC2", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Compute Instance"),
		termMatch("operatingSystem", "Linux"),
		termMatch("tenancy", "Shared"),
		termMatch("preInstalledSw", "NA"),
		termMatch("capacitystatus", "Used"),
	}, func(attrs map[string]string) (string, bool) {
		return attrs["instanceType"], attrs["instanceType"] != ""
	})

	fetch(categorySecretsManager, "AWSSecretsManager", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Secret"),
	}, matchUsagetypeContains("SecretsManager-Secrets"))

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

func (s *service) CalculateECRStorageMonthlyCost(sizeGB int64) float64 {
	if v, ok := s.lookupMonthly(priceKey(categoryECR, ""), 0); ok {
		return float64(sizeGB) * v
	}

	return float64(sizeGB) * ECRStorageCostPerGBMonth
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
		total += s.sagemakerInstanceCost(v.InstanceType) * float64(v.InstanceCount)
	}

	return total
}

func (s *service) CalculateEC2InstanceMonthlyCost(instanceType string) float64 {
	if v, ok := s.lookupMonthly(priceKey(categoryEC2Instance, instanceType), hoursPerMonth); ok {
		return v
	}

	return ec2InstancePricing[instanceType]
}

func (s *service) CalculateSecretsManagerMonthlyCost(count int) float64 {
	if v, ok := s.lookupMonthly(priceKey(categorySecretsManager, ""), 0); ok {
		return float64(count) * v
	}

	return float64(count) * SecretsManagerCostPerSecretMonth
}

func (s *service) sagemakerInstanceCost(instanceType string) float64 {
	if v, ok := s.lookupMonthly(priceKey(categorySageMakerHosting, instanceType), hoursPerMonth); ok {
		return v
	}

	return sagemakerInstancePricing[instanceType]
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
	var doc priceListDocument
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
