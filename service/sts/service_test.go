package awssts

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg)
	assert.NotNil(t, svc)
}

func TestGetCallerIdentity(t *testing.T) {
	mockClient := new(awsinterfaces.MockSTSClient)
	s := &service{client: mockClient}

	expectedOutput := &sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
		Arn:     aws.String("arn:aws:iam::123456789012:user/test"),
		UserId:  aws.String("AKIAI44QH8DHBEXAMPLE"),
	}

	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything, mock.Anything).Return(expectedOutput, nil)

	output, err := s.GetCallerIdentity(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
	mockClient.AssertExpectations(t)
}

func TestGetCallerIdentity_Error(t *testing.T) {
	mockClient := new(awsinterfaces.MockSTSClient)
	s := &service{client: mockClient}

	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("STS error"))

	output, err := s.GetCallerIdentity(context.Background())

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "STS error", err.Error())
	mockClient.AssertExpectations(t)
}
