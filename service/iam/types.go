package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
)

// Service defines the interface for the IAM service.
type Service interface {
	analyzer.WasteAnalyzer
	// GetIAMWaste returns a list of unused IAM users and an alert if the root account lacks MFA.
	GetIAMWaste(ctx context.Context, idleDays int) ([]model.IAMUserWasteInfo, []model.IAMRootUserWasteInfo, error)
}

// ClientAPI defines the AWS IAM client methods used by the service.
type ClientAPI interface {
	ListUsers(ctx context.Context, params *iam.ListUsersInput, optFns ...func(*iam.Options)) (*iam.ListUsersOutput, error)
	ListAccessKeys(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error)
	GetAccessKeyLastUsed(ctx context.Context, params *iam.GetAccessKeyLastUsedInput, optFns ...func(*iam.Options)) (*iam.GetAccessKeyLastUsedOutput, error)
	GetAccountSummary(ctx context.Context, params *iam.GetAccountSummaryInput, optFns ...func(*iam.Options)) (*iam.GetAccountSummaryOutput, error)
}
