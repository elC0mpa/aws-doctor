// Package pricing loader: fetches region-aware prices from the AWS Pricing API at startup and
// caches them in memory for the process lifetime. Lookup functions check this cache first and
// fall back to the hardcoded us-east-1 constants when an entry is missing or Load was never
// called (e.g., unit tests).
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
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"golang.org/x/sync/errgroup"
)

// The Pricing API endpoint is only served from us-east-1 and ap-south-1, but it returns pricing
// data for every AWS region via the regionCode filter on each query. We always talk to us-east-1
// since it is available to every AWS account.
const pricingEndpointRegion = "us-east-1"

// maxPricingConcurrency caps parallel GetProducts calls during Load.
const maxPricingConcurrency = 8

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

// prices is the in-memory pricing cache. It lives for the process lifetime: Load populates it
// once at startup (before any Calculate* call) and no code path invalidates or refreshes it.
// This is intentional — aws-doctor is a short-lived CLI invocation, so a single snapshot is
// sufficient and avoids re-querying the Pricing API. If aws-doctor ever grows into a long-running
// process, this cache will need a TTL or explicit invalidation hook.
//
//nolint:gochecknoglobals // runtime price cache
var (
	priceMu sync.RWMutex
	prices  = map[string]float64{}
)

// priceKey builds a cache key; variant may be empty.
func priceKey(category, variant string) string {
	if variant == "" {
		return category
	}

	return category + ":" + variant
}

// clientAPI is the subset of the Pricing API used by Load.
type clientAPI interface {
	GetProducts(ctx context.Context, params *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error)
}

// Load populates the in-memory price cache for the given AWS region. Partial failures are
// tolerated (missing entries fall back to the hardcoded defaults), but any Pricing API errors
// encountered are joined and returned so the caller can surface them. awsconfig is cloned with
// the Pricing API endpoint region so Load works regardless of which region the caller is using.
func Load(ctx context.Context, awsconfig aws.Config) error {
	cfg := awsconfig.Copy()
	cfg.Region = pricingEndpointRegion

	client := pricing.NewFromConfig(cfg)

	return loadWithClient(ctx, client, awsconfig.Region)
}

// loadWithClient is the testable entry point: it accepts a Pricing API client interface and the
// target AWS region whose prices should be fetched.
func loadWithClient(ctx context.Context, client clientAPI, region string) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxPricingConcurrency)

	var (
		errMu    sync.Mutex
		fetchErr []error
	)

	fetch := func(category, serviceCode string, filters []pricingtypes.Filter, extract func(attrs map[string]string) (variant string, ok bool)) {
		g.Go(func() error {
			entries, err := queryProducts(ctx, client, serviceCode, filters)
			if err != nil {
				wrapped := fmt.Errorf("%s: %w", category, err)

				errMu.Lock()
				defer errMu.Unlock()

				fetchErr = append(fetchErr, wrapped)

				return nil
			}

			for _, entry := range entries {
				variant, ok := extract(entry.attrs)
				if !ok {
					continue
				}

				setPrice(priceKey(category, variant), entry.pricePerUnit)
			}

			return nil
		})
	}

	regionFilter := pricingtypes.Filter{
		Type:  pricingtypes.FilterTypeTermMatch,
		Field: aws.String("regionCode"),
		Value: aws.String(region),
	}

	// matchUsagetypeContains returns an extractor that accepts rows whose usagetype contains the
	// given substring. Pricing API usagetypes are region-prefixed in some services and not in
	// others, so substring matching is more reliable than suffix matching.
	matchUsagetypeContains := func(needle string) func(attrs map[string]string) (string, bool) {
		return func(attrs map[string]string) (string, bool) {
			if !strings.Contains(attrs["usagetype"], needle) {
				return "", false
			}

			return "", true
		}
	}

	// matchLBUsage accepts standard "LoadBalancerUsage" rows while rejecting "Outposts-" and
	// "TS-" variants (which use the same productFamily but distinct pricing). The usagetype has
	// an optional region prefix across both ALB and CLB product families.
	matchLBUsage := func(attrs map[string]string) (string, bool) {
		u := attrs["usagetype"]
		if !strings.Contains(u, "LoadBalancerUsage") {
			return "", false
		}

		if strings.Contains(u, "Outposts-") || strings.Contains(u, "TS-") {
			return "", false
		}

		return "", true
	}

	// EBS: one filter per volumeApiName.
	for _, v := range []string{"gp2", "gp3", "io1", "io2", "st1", "sc1"} {
		fetch(categoryEBS, "AmazonEC2", []pricingtypes.Filter{
			regionFilter,
			termMatch("productFamily", "Storage"),
			termMatch("volumeApiName", v),
		}, func(attrs map[string]string) (string, bool) {
			return attrs["volumeApiName"], attrs["volumeApiName"] != ""
		})
	}

	// Elastic IP: unassociated (idle) rate lives in AmazonVPC, group=VPCPublicIPv4Address,
	// usagetype contains "PublicIPv4:IdleAddress".
	fetch(categoryEIP, "AmazonVPC", []pricingtypes.Filter{
		regionFilter,
		termMatch("group", "VPCPublicIPv4Address"),
	}, matchUsagetypeContains("PublicIPv4:IdleAddress"))

	// NAT Gateway hourly rate: usagetype contains "NatGateway-Hours" to skip data/bytes rates.
	fetch(categoryNAT, "AmazonEC2", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "NAT Gateway"),
		termMatch("operation", "NatGateway"),
	}, matchUsagetypeContains("NatGateway-Hours"))

	// Application and Network load balancers live under distinct productFamily values. Both
	// typically price at the same $0.0225/hr rate, but Network LBs are queried separately to
	// make the lookup robust if the rates ever diverge regionally. The matchLBUsage extractor
	// drops LCU, TS-, and Outposts-* variants that share the family.
	for _, family := range []string{"Load Balancer-Application", "Load Balancer-Network"} {
		fetch(categoryLBApp, "AWSELB", []pricingtypes.Filter{
			regionFilter,
			termMatch("productFamily", family),
		}, matchLBUsage)
	}

	// Classic load balancer: same matcher under productFamily "Load Balancer".
	fetch(categoryLBClassic, "AWSELB", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Load Balancer"),
	}, matchLBUsage)

	// CloudWatch Logs storage: usagetype contains "TimedStorage-ByteHrs" (pricing is per GB-Mo).
	fetch(categoryCWLogs, "AmazonCloudWatch", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Storage Snapshot"),
	}, matchUsagetypeContains("TimedStorage-ByteHrs"))

	// RDS General Purpose gp2 storage: usagetype contains "RDS:GP2-Storage" (distinct from
	// "RDSCustom:..." and "Multi-AZ-GP2-Storage" variants).
	fetch(categoryRDSStorage, "AmazonRDS", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Database Storage"),
		termMatch("volumeType", "General Purpose"),
		termMatch("deploymentOption", "Single-AZ"),
	}, matchUsagetypeContains("RDS:GP2-Storage"))

	// RDS snapshot storage: usagetype contains "RDS:ChargedBackupUsage", which excludes
	// "RDSCustom:ChargedBackupUsage" variants.
	fetch(categoryRDSSnapshot, "AmazonRDS", []pricingtypes.Filter{
		regionFilter,
		termMatch("productFamily", "Storage Snapshot"),
	}, matchUsagetypeContains("RDS:ChargedBackupUsage"))

	// RDS instance compute: MySQL On-Demand Single-AZ is used as a proxy (rates are close across
	// engines for estimation purposes and match how the hardcoded fallback table was built).
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

// productEntry captures the parsed essentials of one PriceList JSON document.
type productEntry struct {
	attrs        map[string]string
	pricePerUnit float64
}

// queryProducts runs a paginated GetProducts request and parses each returned price list JSON
// document into a productEntry containing the on-demand hourly/monthly USD rate.
func queryProducts(ctx context.Context, client clientAPI, serviceCode string, filters []pricingtypes.Filter) ([]productEntry, error) {
	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String(serviceCode),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
	}

	paginator := pricing.NewGetProductsPaginator(client, input)

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

// parsePriceListDocument pulls product attributes and the first on-demand USD price per unit out
// of a single Pricing API JSON document.
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

	// Take the first price dimension found. This is correct for the flat-rate categories we
	// currently query (EBS, NAT, LB, CW Logs, RDS storage/snapshot, RDS instance) which each
	// have a single "per unit" dimension. Tiered products (e.g., S3 storage or data transfer)
	// would require inspecting beginRange/endRange on each dimension.
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

func termMatch(field, value string) pricingtypes.Filter {
	return pricingtypes.Filter{
		Type:  pricingtypes.FilterTypeTermMatch,
		Field: aws.String(field),
		Value: aws.String(value),
	}
}

func setPrice(k string, v float64) {
	priceMu.Lock()
	defer priceMu.Unlock()

	prices[k] = v
}

// lookupMonthly returns the cached monthly rate for k if present. Pricing API rates are hourly
// for compute and per-GB-month for storage; the caller tells us which by passing hoursInMonth.
// When the cached value is absent, ok is false and the caller falls back to constants.
func lookupMonthly(k string, hoursInMonth float64) (float64, bool) {
	priceMu.RLock()
	defer priceMu.RUnlock()

	v, ok := prices[k]
	if !ok {
		return 0, false
	}

	if hoursInMonth > 0 {
		return v * hoursInMonth, true
	}

	return v, true
}

// resetForTest clears the cache between test runs. Only intended for use in tests within this
// package.
func resetForTest() {
	priceMu.Lock()
	defer priceMu.Unlock()

	prices = map[string]float64{}
}
