package pricing

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	awspricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the subset of the AWS Pricing API used by the service.
type ClientAPI interface {
	GetProducts(ctx context.Context, params *awspricing.GetProductsInput, optFns ...func(*awspricing.Options)) (*awspricing.GetProductsOutput, error)
}

// Service is the interface for AWS pricing estimates.
type Service interface {
	LoadRegionRates(ctx context.Context) error
	CalculateEBSMonthlyCost(sizeGiB int32, volumeType types.VolumeType) float64
	CalculateEBSSnapshotMonthlyCost(sizeGB int64) float64
	CalculateEIPMonthlyCost() float64
	CalculateLoadBalancerMonthlyCost(lbType elbtypes.LoadBalancerTypeEnum) float64
	CalculateCloudWatchLogsMonthlyCost(storedBytes int64) float64
	CalculateNATGatewayMonthlyCost() float64
	CalculateRDSInstanceMonthlyCost(allocatedGB int32, multiAZ bool) float64
	CalculateRDSSnapshotMonthlyCost(allocatedGB int32) float64
	CalculateRDSIdleInstanceMonthlyCost(instanceClass string, allocatedGB int32, multiAZ bool) float64
	CalculateSageMakerEndpointMonthlyCost(variants []model.SageMakerVariant) float64
	CalculateECRStorageMonthlyCost(sizeGB int64) float64
	CalculateSecretsManagerMonthlyCost(count int) float64
}

type service struct {
	client  ClientAPI
	priceMu sync.RWMutex
	prices  map[string]float64
	region  string
}

// Internal structs for unmarshaling Pricing API JSON responses.

type priceListDimension struct {
	PricePerUnit map[string]string `json:"pricePerUnit"`
}

type priceListOnDemand struct {
	PriceDimensions map[string]priceListDimension `json:"priceDimensions"`
}

type priceListTerms struct {
	OnDemand map[string]priceListOnDemand `json:"OnDemand"`
}

type priceListProduct struct {
	Attributes map[string]string `json:"attributes"`
}

type priceListDocument struct {
	Product priceListProduct `json:"product"`
	Terms   priceListTerms   `json:"terms"`
}
