package awssts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// STSClientAPI is the interface for the AWS STS client methods used by the service.
type STSClientAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type service struct {
	client STSClientAPI
}

// Service is the interface for AWS STS service.
type Service interface {
	GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error)
}
