package s3

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBucketsWithoutLifecyclePolicies(t *testing.T) {
	ctx := context.Background()
	creationDate := time.Now()

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockS3Client)
		wantCount  int
		wantErr    bool
	}{
		{
			name: "bucket without lifecycle policy",
			setupMocks: func(m *awsinterfaces.MockS3Client) {
				m.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListBucketsOutput{
					Buckets: []types.Bucket{
						{Name: aws.String("no-policy-bucket"), CreationDate: &creationDate},
					},
				}, nil)
				m.On("GetBucketLifecycleConfiguration", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetBucketLifecycleConfigurationOutput)(nil), &smithy.GenericAPIError{
					Code: "NoSuchLifecycleConfiguration",
				})
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "bucket with lifecycle policy",
			setupMocks: func(m *awsinterfaces.MockS3Client) {
				m.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListBucketsOutput{
					Buckets: []types.Bucket{
						{Name: aws.String("with-policy-bucket"), CreationDate: &creationDate},
					},
				}, nil)
				m.On("GetBucketLifecycleConfiguration", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetBucketLifecycleConfigurationOutput{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "list buckets fails",
			setupMocks: func(m *awsinterfaces.MockS3Client) {
				m.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return((*s3.ListBucketsOutput)(nil), errors.New("list error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockS3Client)
			tt.setupMocks(mockClient)

			svc := &service{client: mockClient}
			results, err := svc.GetBucketsWithoutLifecyclePolicies(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.wantCount)

				if tt.wantCount > 0 {
					assert.Equal(t, "no-policy-bucket", results[0].BucketName)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}
