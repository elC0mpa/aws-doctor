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

func TestGetS3Waste(t *testing.T) {
	ctx := context.Background()
	creationDate := time.Now()

	tests := []struct {
		name               string
		setupMocks         func(*awsinterfaces.MockS3Client)
		wantBucketsCount   int
		wantMultipartCount int
		wantErr            bool
	}{
		{
			name: "bucket with both types of waste",
			setupMocks: func(m *awsinterfaces.MockS3Client) {
				m.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListBucketsOutput{
					Buckets: []types.Bucket{
						{Name: aws.String("waste-bucket"), CreationDate: &creationDate},
					},
				}, nil)
				// Lifecycle check
				m.On("GetBucketLifecycleConfiguration", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetBucketLifecycleConfigurationOutput)(nil), &smithy.GenericAPIError{
					Code: "NoSuchLifecycleConfiguration",
				})
				// Multipart check
				m.On("ListMultipartUploads", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListMultipartUploadsOutput{
					Uploads: []types.MultipartUpload{
						{UploadId: aws.String("upload-1")},
					},
				}, nil)
			},
			wantBucketsCount:   1,
			wantMultipartCount: 1,
			wantErr:            false,
		},
		{
			name: "clean bucket",
			setupMocks: func(m *awsinterfaces.MockS3Client) {
				m.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListBucketsOutput{
					Buckets: []types.Bucket{
						{Name: aws.String("clean-bucket"), CreationDate: &creationDate},
					},
				}, nil)
				// Lifecycle check - success means it has policy
				m.On("GetBucketLifecycleConfiguration", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetBucketLifecycleConfigurationOutput{}, nil)
				// Multipart check - empty means no waste
				m.On("ListMultipartUploads", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListMultipartUploadsOutput{
					Uploads: []types.MultipartUpload{},
				}, nil)
			},
			wantBucketsCount:   0,
			wantMultipartCount: 0,
			wantErr:            false,
		},
		{
			name: "list buckets fails",
			setupMocks: func(m *awsinterfaces.MockS3Client) {
				m.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return((*s3.ListBucketsOutput)(nil), errors.New("list error"))
			},
			wantBucketsCount:   0,
			wantMultipartCount: 0,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockS3Client)
			tt.setupMocks(mockClient)

			svc := &service{client: mockClient}
			buckets, multiparts, err := svc.GetS3Waste(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, buckets, tt.wantBucketsCount)
				assert.Len(t, multiparts, tt.wantMultipartCount)

				if tt.wantBucketsCount > 0 {
					assert.Equal(t, "waste-bucket", buckets[0].BucketName)
				}

				if tt.wantMultipartCount > 0 {
					assert.Equal(t, "waste-bucket", multiparts[0].BucketName)
					assert.Equal(t, 1, multiparts[0].UploadCount)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}
