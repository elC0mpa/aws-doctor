// Package awssts provides a service for interacting with AWS STS.
package awssts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// NewService creates a new STS service.
func NewService(awsconfig aws.Config) Service {
	client := sts.NewFromConfig(awsconfig)

	return &service{
		client: client,
	}
}

func (s *service) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	input := &sts.GetCallerIdentityInput{}

	return s.client.GetCallerIdentity(ctx, input)
}
