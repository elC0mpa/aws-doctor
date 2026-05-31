package secretsmanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	s := NewService(aws.Config{}, nil)
	assert.NotNil(t, s)
}

func TestGetUnusedSecrets(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	idleDays := 90
	threshold := now.AddDate(0, 0, -idleDays)

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockSecretsManagerClient)
		wantCount  int
		wantErr    bool
	}{
		{
			name: "mixed secrets",
			setupMocks: func(m *awsinterfaces.MockSecretsManagerClient) {
				m.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Return(&secretsmanager.ListSecretsOutput{
					SecretList: []types.SecretListEntry{
						{
							Name:             aws.String("unused-1"),
							LastAccessedDate: aws.Time(threshold.Add(-time.Hour)),
							PrimaryRegion:    aws.String("us-east-1"),
						},
						{
							Name:             aws.String("never-accessed"),
							LastAccessedDate: nil,
							PrimaryRegion:    aws.String("us-east-1"),
						},
						{
							Name:             aws.String("recent"),
							LastAccessedDate: aws.Time(now),
							PrimaryRegion:    aws.String("us-east-1"),
						},
						{
							Name:             aws.String("replica"),
							LastAccessedDate: nil,
							PrimaryRegion:    aws.String("us-west-2"),
						},
					},
				}, nil)
			},
			wantCount: 2, // unused-1, never-accessed
			wantErr:   false,
		},
		{
			name: "list error",
			setupMocks: func(m *awsinterfaces.MockSecretsManagerClient) {
				m.On("ListSecrets", mock.Anything, mock.Anything, mock.Anything).Return((*secretsmanager.ListSecretsOutput)(nil), errors.New("list error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockSecretsManagerClient)
			tt.setupMocks(mockClient)

			svc := &service{
				client:        mockClient,
				currentRegion: "us-east-1",
			}

			secrets, err := svc.GetUnusedSecrets(ctx, idleDays)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, secrets, tt.wantCount)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAnalyzerMethods(t *testing.T) {
	svc := &service{}
	
	if svc.Name() == "" {
		t.Error("Name() should not be empty")
	}
	
	if svc.TabName() == "" {
		t.Error("TabName() should not be empty")
	}
}
