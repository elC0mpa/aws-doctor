package awssts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type service struct {
	client *sts.Client
}

// Service is the interface for AWS STS service.
type Service interface {
	GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error)
}
