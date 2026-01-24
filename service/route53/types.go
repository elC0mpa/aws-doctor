package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/elC0mpa/aws-doctor/model"
)

type service struct {
	client *route53.Client
}

// Route53Service defines the interface for Route 53 waste detection
type Route53Service interface {
	GetEmptyHostedZones(ctx context.Context) ([]model.HostedZoneWasteInfo, error)
}
