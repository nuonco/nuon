package helpers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

func (h *Helpers) EnsureInstallRunnerCloudIdentity(ctx context.Context, install *app.Install, runnerID string) (*app.RunnerCloudIdentity, error) {
	accountID := install.CloudPlatformMetadata.TargetAccountID
	if accountID == "" {
		return nil, nil
	}

	ctx = cctx.SetOrgIDContext(ctx, install.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, install.CreatedByID)

	awsAccount := install.AWSAccount
	if awsAccount == nil {
		var loaded app.AWSAccount
		err := h.db.WithContext(ctx).Where(app.AWSAccount{InstallID: install.ID}).First(&loaded).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("unable to get install aws account: %w", err)
		}
		awsAccount = &loaded
	}
	if awsAccount.Region == "" {
		return nil, nil
	}

	existing, err := h.getRunnerCloudIdentity(ctx, runnerID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to get runner cloud identity: %w", err)
	}

	identity := app.RunnerCloudIdentity{
		RunnerID:  runnerID,
		Provider:  app.CloudPlatformAWS,
		AccountID: accountID,
		Region:    awsAccount.Region,
	}
	if err := h.db.WithContext(ctx).Create(&identity).Error; err != nil {
		if request.IsDuplicateKey(err) {
			return h.getRunnerCloudIdentity(ctx, runnerID)
		}
		return nil, fmt.Errorf("unable to create runner cloud identity: %w", err)
	}

	return &identity, nil
}

func (h *Helpers) getRunnerCloudIdentity(ctx context.Context, runnerID string) (*app.RunnerCloudIdentity, error) {
	var identity app.RunnerCloudIdentity
	if err := h.db.WithContext(ctx).
		Where(app.RunnerCloudIdentity{RunnerID: runnerID}).
		First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

type RunnerCloudIdentityRoleUpdate struct {
	Bound           bool   `json:"bound"`
	Accepted        bool   `json:"accepted"`
	Reason          string `json:"reason,omitempty"`
	ExpectedAccount string `json:"expected_account,omitempty"`
}

func (h *Helpers) SetRunnerCloudIdentityRole(ctx context.Context, runnerID, reportedRoleARN string) (*RunnerCloudIdentityRoleUpdate, error) {
	identity, err := h.getRunnerCloudIdentity(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &RunnerCloudIdentityRoleUpdate{}, nil
		}
		return nil, fmt.Errorf("unable to get runner cloud identity: %w", err)
	}

	update := &RunnerCloudIdentityRoleUpdate{Bound: true, ExpectedAccount: identity.AccountID}
	if reportedRoleARN == "" {
		update.Reason = "no runner role reported"
		return update, nil
	}
	parsed, err := arn.Parse(reportedRoleARN)
	if err != nil || parsed.Service != "iam" || !strings.HasPrefix(parsed.Resource, "role/") {
		update.Reason = "reported runner role is not an IAM role ARN"
		return update, nil
	}
	if parsed.AccountID != identity.AccountID {
		update.Reason = "reported runner role is in another account"
		return update, nil
	}

	update.Accepted = true
	if identity.RunnerRoleARN == reportedRoleARN {
		return update, nil
	}
	if err := h.db.WithContext(ctx).
		Model(&app.RunnerCloudIdentity{}).
		Where(app.RunnerCloudIdentity{ID: identity.ID}).
		Updates(app.RunnerCloudIdentity{RunnerRoleARN: reportedRoleARN}).Error; err != nil {
		return nil, fmt.Errorf("unable to update runner cloud identity: %w", err)
	}

	return update, nil
}
