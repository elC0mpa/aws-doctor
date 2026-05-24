package iam

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

type service struct {
	client ClientAPI
}

// NewService creates a new instance of the IAM service.
func NewService(awsconfig aws.Config) Service {
	client := iam.NewFromConfig(awsconfig)
	return &service{client: client}
}

func (s *service) GetIAMWaste(ctx context.Context, idleDays int) ([]model.IAMUserWasteInfo, []model.IAMRootUserWasteInfo, error) {
	var rootWaste []model.IAMRootUserWasteInfo

	// 1. Root MFA check
	summaryOutput, err := s.client.GetAccountSummary(ctx, &iam.GetAccountSummaryInput{})
	if err != nil {
		return nil, nil, fmt.Errorf("getting account summary: %w", err)
	}

	if val, ok := summaryOutput.SummaryMap["AccountMFAEnabled"]; ok && val == 0 {
		rootWaste = append(rootWaste, model.IAMRootUserWasteInfo{HasMFA: false})
	}

	// 2. IAM Users check
	var (
		users []model.IAMUserWasteInfo
		mu    sync.Mutex
	)

	cutoffTime := time.Now().UTC().AddDate(0, 0, -idleDays)
	paginator := iam.NewListUsersPaginator(s.client, &iam.ListUsersInput{})

	g, groupCtx := errgroup.WithContext(ctx)
	// We set a reasonable concurrency limit to avoid throttling
	g.SetLimit(10)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(groupCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("listing IAM users: %w", err)
		}

		for _, u := range page.Users {
			user := u // Capture variable for closure

			g.Go(func() error {
				wasteInfo, err := s.processIAMUser(groupCtx, user, cutoffTime, idleDays)
				if err != nil {
					return err
				}

				if wasteInfo != nil {
					mu.Lock()

					users = append(users, *wasteInfo)
					mu.Unlock()
				}

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	return users, rootWaste, nil
}

func (s *service) processIAMUser(ctx context.Context, user iamtypes.User, cutoffTime time.Time, idleDays int) (*model.IAMUserWasteInfo, error) {
	// Check PasswordLastUsed
	if user.PasswordLastUsed != nil && user.PasswordLastUsed.After(cutoffTime) {
		return nil, nil // Password used recently, not waste
	}

	// If never used, but created recently, skip it
	if user.PasswordLastUsed == nil && user.CreateDate != nil && user.CreateDate.After(cutoffTime) {
		return nil, nil
	}

	// Check access keys
	keysOutput, err := s.client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
		UserName: user.UserName,
	})
	if err != nil {
		return nil, fmt.Errorf("listing access keys for %s: %w", aws.ToString(user.UserName), err)
	}

	activeKeyFound := false

	for _, key := range keysOutput.AccessKeyMetadata {
		lastUsedOut, err := s.client.GetAccessKeyLastUsed(ctx, &iam.GetAccessKeyLastUsedInput{
			AccessKeyId: key.AccessKeyId,
		})
		if err != nil {
			return nil, fmt.Errorf("getting access key last used for %s: %w", aws.ToString(key.AccessKeyId), err)
		}

		if lastUsedOut.AccessKeyLastUsed != nil && lastUsedOut.AccessKeyLastUsed.LastUsedDate != nil {
			if lastUsedOut.AccessKeyLastUsed.LastUsedDate.After(cutoffTime) {
				activeKeyFound = true
				break
			}
		}
	}

	if activeKeyFound {
		return nil, nil
	}

	// Construct waste info
	pwdStatus := "Never"

	if user.PasswordLastUsed != nil {
		days := int(time.Since(*user.PasswordLastUsed).Hours() / 24)
		pwdStatus = fmt.Sprintf("%d days ago", days)
	}

	keysStatus := "No access keys"
	if len(keysOutput.AccessKeyMetadata) > 0 {
		keysStatus = fmt.Sprintf("All %d keys unused > %d days", len(keysOutput.AccessKeyMetadata), idleDays)
	}

	return &model.IAMUserWasteInfo{
		UserName:         aws.ToString(user.UserName),
		PasswordLastUsed: pwdStatus,
		AccessKeysStatus: keysStatus,
	}, nil
}
