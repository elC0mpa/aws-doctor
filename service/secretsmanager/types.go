package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS Secrets Manager client methods used by the service.
type ClientAPI interface {
	ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
}

type service struct {
	client         ClientAPI
	currentRegion  string
	pricingService pricingService
}

// Service is the interface for AWS Secrets Manager service.
type Service interface {
	GetUnusedSecrets(ctx context.Context, idleDays int) ([]model.UnusedSecretInfo, error)
}

// pricingService is a local interface for the pricing dependency.
type pricingService interface {
	CalculateSecretsManagerMonthlyCost(count int) float64
}
