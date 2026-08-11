package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type CreateAppInstallConfigSyncInput struct {
	AppID         string `json:"app_id" validate:"required"`
	CommitSHA     string `json:"commit_sha,omitempty"`
	TriggeredBy   string `json:"triggered_by"`
	CreatedByID   string `json:"created_by_id,omitempty"`
	QueueSignalID string `json:"queue_signal_id,omitempty"`
}

type CreateAppInstallConfigSyncOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateAppInstallConfigSync(ctx context.Context, input *CreateAppInstallConfigSyncInput) (*CreateAppInstallConfigSyncOutput, error) {
	createdByID := keys.CreatedByIDFromContext(ctx)
	if createdByID == "" && input.CreatedByID != "" {
		createdByID = input.CreatedByID
	}
	if createdByID == "" {
		panic(fmt.Sprintf("CreateAppInstallConfigSync: created_by_id is empty - context keys: account_id=%v, org_id=%v",
			ctx.Value(keys.AccountIDCtxKey), ctx.Value(keys.OrgIDCtxKey)))
	}

	record := app.AppInstallConfigSync{
		AppID:         input.AppID,
		CreatedByID:   createdByID,
		TriggeredBy:   input.TriggeredBy,
		QueueSignalID: input.QueueSignalID,
		Status:        app.NewCompositeStatus(ctx, app.StatusInProgress),
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
		return nil, fmt.Errorf("unable to create app install config sync (created_by=%s, org=%s, app=%s): %w",
			createdByID, keys.OrgIDFromContext(ctx), input.AppID, err)
	}

	return &CreateAppInstallConfigSyncOutput{ID: record.ID}, nil
}
