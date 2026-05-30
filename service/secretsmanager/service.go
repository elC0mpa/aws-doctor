package secretsmanager

import (
	"context"
	"errors"
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

func (s *service) Name() string {
	return "secrets-manager"
}

func (s *service) TabName() string {
	return "SecretsManager"
}

func (s *service) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	start := time.Now()
	input := model.RenderWasteInput{}

	var errs []error

	idleDays := flags.SecretsIdleDays
	if idleDays == 0 {
		idleDays = 90
	}

	unusedSecrets, err := s.GetUnusedSecrets(ctx, idleDays)
	if err != nil {
		errs = append(errs, err)
	} else {
		input.UnusedSecrets = unusedSecrets
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("secrets-manager analyze errors: %w", errors.Join(errs...))
	}

	return model.ScopeResult{
		Scope:    s.Name(),
		Input:    input,
		Duration: time.Since(start),
		Err:      finalErr,
	}, nil
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
