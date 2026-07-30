package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateAppInstallConfigSyncInput struct {
	AppID       string `json:"app_id" validate:"required"`
	CommitSHA   string `json:"commit_sha,omitempty"`
	TriggeredBy string `json:"triggered_by"`
}

type CreateAppInstallConfigSyncOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateAppInstallConfigSync(ctx context.Context, input *CreateAppInstallConfigSyncInput) (*CreateAppInstallConfigSyncOutput, error) {
	record := app.AppInstallConfigSync{
		AppID:       input.AppID,
		TriggeredBy: input.TriggeredBy,
		Status:      app.NewCompositeStatus(ctx, app.StatusInProgress),
	}

	if input.CommitSHA != "" {
		commit := app.VCSConnectionCommit{
			SHA: input.CommitSHA,
		}
		if err := a.db.WithContext(ctx).Create(&commit).Error; err == nil {
			record.VCSConnectionCommitID = &commit.ID
		}
	}

	if err := a.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("unable to create app install config sync: %w", err)
	}

	return &CreateAppInstallConfigSyncOutput{ID: record.ID}, nil
}
