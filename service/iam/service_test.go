package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIAMWaste(t *testing.T) {
	now := time.Now().UTC()
	recent := now.AddDate(0, 0, -10)
	old := now.AddDate(0, 0, -100)

	tests := []struct {
		name      string
		idleDays  int
		setupMock func(*awsinterfaces.MockIAMClientAPI)
		wantUsers []model.IAMUserWasteInfo
		wantRoot  []model.IAMRootUserWasteInfo
		wantErr   bool
	}{
		{
			name:     "no waste found",
			idleDays: 90,
			setupMock: func(m *awsinterfaces.MockIAMClientAPI) {
				m.On("GetAccountSummary", mock.Anything, &iam.GetAccountSummaryInput{}).Return(&iam.GetAccountSummaryOutput{
					SummaryMap: map[string]int32{"AccountMFAEnabled": 1},
				}, nil)

				m.On("ListUsers", mock.Anything, &iam.ListUsersInput{}).Return(&iam.ListUsersOutput{
					Users: []types.User{
						{UserName: aws.String("active-user"), PasswordLastUsed: &recent},
					},
					IsTruncated: false,
				}, nil)
			},
			wantUsers: nil,
			wantRoot:  nil,
			wantErr:   false,
		},
		{
			name:     "root MFA not enabled and one unused user",
			idleDays: 90,
			setupMock: func(m *awsinterfaces.MockIAMClientAPI) {
				m.On("GetAccountSummary", mock.Anything, &iam.GetAccountSummaryInput{}).Return(&iam.GetAccountSummaryOutput{
					SummaryMap: map[string]int32{"AccountMFAEnabled": 0},
				}, nil)

				m.On("ListUsers", mock.Anything, &iam.ListUsersInput{}).Return(&iam.ListUsersOutput{
					Users: []types.User{
						{UserName: aws.String("idle-user"), PasswordLastUsed: &old, CreateDate: &old},
					},
					IsTruncated: false,
				}, nil)

				m.On("ListAccessKeys", mock.Anything, &iam.ListAccessKeysInput{
					UserName: aws.String("idle-user"),
				}).Return(&iam.ListAccessKeysOutput{
					AccessKeyMetadata: []types.AccessKeyMetadata{
						{AccessKeyId: aws.String("AKIA123")},
					},
				}, nil)

				m.On("GetAccessKeyLastUsed", mock.Anything, &iam.GetAccessKeyLastUsedInput{
					AccessKeyId: aws.String("AKIA123"),
				}).Return(&iam.GetAccessKeyLastUsedOutput{
					AccessKeyLastUsed: &types.AccessKeyLastUsed{
						LastUsedDate: &old,
					},
				}, nil)
			},
			wantUsers: []model.IAMUserWasteInfo{
				{
					UserName:         "idle-user",
					PasswordLastUsed: "100 days ago", // Approximate, logic uses exact time.Since
					AccessKeysStatus: "All 1 keys unused > 90 days",
				},
			},
			wantRoot: []model.IAMRootUserWasteInfo{
				{HasMFA: false},
			},
			wantErr: false,
		},
		{
			name:     "error getting account summary",
			idleDays: 90,
			setupMock: func(m *awsinterfaces.MockIAMClientAPI) {
				m.On("GetAccountSummary", mock.Anything, &iam.GetAccountSummaryInput{}).Return(nil, errors.New("api error"))
				m.On("ListUsers", mock.Anything, &iam.ListUsersInput{}).Return(&iam.ListUsersOutput{
					Users:       []types.User{},
					IsTruncated: false,
				}, nil)
			},
			wantUsers: nil,
			wantRoot:  nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockIAMClientAPI)
			tt.setupMock(mockClient)

			svc := &service{client: mockClient}
			gotUsers, gotRoot, err := svc.GetIAMWaste(context.Background(), tt.idleDays)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// We only check lengths and basic data because time.Since will slightly vary
			assert.Equal(t, len(tt.wantUsers), len(gotUsers))

			if len(tt.wantUsers) > 0 {
				assert.Equal(t, tt.wantUsers[0].UserName, gotUsers[0].UserName)
				assert.Equal(t, tt.wantUsers[0].AccessKeysStatus, gotUsers[0].AccessKeysStatus)
			}

			assert.Equal(t, tt.wantRoot, gotRoot)
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

func TestService_Analyze(t *testing.T) {
	mockClient := new(awsinterfaces.MockIAMClientAPI)
	svc := &service{client: mockClient}

	assert.Equal(t, "iam", svc.Name())
	assert.Equal(t, "IAM", svc.TabName())

	mockClient.On("GetAccountSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("err")).Maybe()
	mockClient.On("GetAccountSummary", mock.Anything, mock.Anything).Return(nil, errors.New("err")).Maybe()

	mockClient.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("err")).Maybe()
	mockClient.On("GetLoginProfile", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("err")).Maybe()
	mockClient.On("ListAccessKeys", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("err")).Maybe()
	mockClient.On("ListRoles", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("err")).Maybe()

	res, err := svc.Analyze(context.Background(), model.Flags{})
	assert.NoError(t, err)
	assert.Equal(t, "iam", res.Scope)
}
