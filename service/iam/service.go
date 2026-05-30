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

	return &service{
		client: client,
	}
}

func (s *service) Name() string {
	return "iam"
}

func (s *service) TabName() string {
	return "IAM"
}

func (s *service) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	start := time.Now()
	input := model.RenderWasteInput{}

	var errs []error

	idleDays := flags.IAMIdleDays
	if idleDays == 0 {
		idleDays = 90
	}

	unusedUsers, rootWaste, err := s.GetIAMWaste(ctx, idleDays)
	if err != nil {
		errs = append(errs, err)
	} else {
		input.UnusedIAMUsers = unusedUsers
		input.RootUserWaste = rootWaste
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("iam analyze errors: %v", errs)
	}

	return model.ScopeResult{
		Scope:    s.Name(),
		Input:    input,
		Duration: time.Since(start),
		Err:      finalErr,
	}, nil
}

func (s *service) GetIAMWaste(ctx context.Context, idleDays int) ([]model.IAMUserWasteInfo, []model.IAMRootUserWasteInfo, error) {
	var rootWaste []model.IAMRootUserWasteInfo

	// 1. Root MFA check
	summaryOutput, err := s.client.GetAccountSummary(ctx, &iam.GetAccountSummaryInput{})
	if err == nil {
		if val, ok := summaryOutput.SummaryMap["AccountMFAEnabled"]; ok && val == 0 {
			rootWaste = append(rootWaste, model.IAMRootUserWasteInfo{HasMFA: false})
		}
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
	paginator := iam.NewListAccessKeysPaginator(s.client, &iam.ListAccessKeysInput{
		UserName: user.UserName,
	})

	var (
		activeKeyFound bool
		keyCount       int
	)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing access keys for %s: %w", aws.ToString(user.UserName), err)
		}

		for _, key := range page.AccessKeyMetadata {
			keyCount++

			lastUsedOut, err := s.client.GetAccessKeyLastUsed(ctx, &iam.GetAccessKeyLastUsedInput{
				AccessKeyId: key.AccessKeyId,
			})
			if err != nil {
				// On error, assume active to avoid false positives in cleanup reports
				activeKeyFound = true
				break
			}

			if lastUsedOut.AccessKeyLastUsed != nil && lastUsedOut.AccessKeyLastUsed.LastUsedDate != nil {
				if lastUsedOut.AccessKeyLastUsed.LastUsedDate.After(cutoffTime) {
					activeKeyFound = true
					break
				}
			} else if key.CreateDate != nil && key.CreateDate.After(cutoffTime) {
				// Key never used, but created recently
				activeKeyFound = true
				break
			}
		}

		if activeKeyFound {
			break
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
	if keyCount > 0 {
		keysStatus = fmt.Sprintf("All %d keys unused > %d days", keyCount, idleDays)
	}

	return &model.IAMUserWasteInfo{
		UserName:         aws.ToString(user.UserName),
		PasswordLastUsed: pwdStatus,
		AccessKeysStatus: keysStatus,
	}, nil
}
