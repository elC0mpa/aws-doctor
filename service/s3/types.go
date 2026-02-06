package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS S3 client methods used by the service.
type ClientAPI interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketLifecycleConfiguration(ctx context.Context, params *s3.GetBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error)
}

type service struct {
	client ClientAPI
}

// Service is the interface for AWS S3 service.
type Service interface {
	GetBucketsWithoutLifecyclePolicies(ctx context.Context) ([]model.S3BucketWasteInfo, error)
}
