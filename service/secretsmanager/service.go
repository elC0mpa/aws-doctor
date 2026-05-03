package secretsmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/elC0mpa/aws-doctor/model"
)

// NewService creates a new Secrets Manager service.
func NewService(awsconfig aws.Config, pricingService pricingService) Service {
	client := secretsmanager.NewFromConfig(awsconfig)

	return &service{
		client:         client,
		currentRegion:  awsconfig.Region,
		pricingService: pricingService,
	}
}

func (s *service) GetUnusedSecrets(ctx context.Context, idleDays int) ([]model.UnusedSecretInfo, error) {
	var unusedSecrets []model.UnusedSecretInfo

	threshold := time.Now().AddDate(0, 0, -idleDays)

	paginator := secretsmanager.NewListSecretsPaginator(s.client, &secretsmanager.ListSecretsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, secret := range output.SecretList {
			// Skip secrets whose PrimaryRegion differs from the current region to avoid double-counting replica secrets.
			if secret.PrimaryRegion != nil && *secret.PrimaryRegion != s.currentRegion {
				continue
			}

			isUnused := false
			if secret.LastAccessedDate == nil {
				isUnused = true
			} else if secret.LastAccessedDate.Before(threshold) {
				isUnused = true
			}

			if isUnused {
				unusedSecrets = append(unusedSecrets, model.UnusedSecretInfo{
					Name:             aws.ToString(secret.Name),
					LastAccessedDate: secret.LastAccessedDate,
				})
			}
		}
	}

	return unusedSecrets, nil
}
