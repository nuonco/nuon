package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"

	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer"
)

type SyncAppConfigInput struct {
	AppConfigID string `json:"app_config_id" validate:"required"`
	AppID       string `json:"app_id" validate:"required"`
	AppBranchID string `json:"app_branch_id" validate:"required"`
}

type SyncAppConfigOutput struct {
	AppConfigID  string   `json:"app_config_id"`
	ComponentIDs []string `json:"component_ids"`
	ActionIDs    []string `json:"action_ids"`
	RunbookIDs   []string `json:"runbook_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) syncAppConfig(ctx context.Context, req *SyncAppConfigInput) (*SyncAppConfigOutput, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Builds are left to the run's builds step, so this sync must not schedule
	// them itself.
	result, err := syncer.Run(ctx, a.syncRunDeps(), syncer.RunRequest{
		AppID:       req.AppID,
		AppConfigID: req.AppConfigID,
	})
	if err != nil {
		var syncErr configsync.SyncErr
		if errors.As(err, &syncErr) {
			return nil, temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("unable to sync config: %s", err.Error()),
				"SYNC_VALIDATION_ERROR",
				err,
			)
		}
		return nil, fmt.Errorf("unable to sync config: %w", err)
	}

	return &SyncAppConfigOutput{
		AppConfigID:  result.AppConfigID,
		ComponentIDs: result.ComponentIDs,
		ActionIDs:    result.ActionIDs,
		RunbookIDs:   result.RunbookIDs,
	}, nil
}

func (a *Activities) syncRunDeps() syncer.RunDeps {
	return syncer.RunDeps{
		DB:               a.db,
		AppsHelpers:      a.helpers,
		ComponentHelpers: a.componentHelpers,
		ActionsHelpers:   a.actionsHelpers,
		RunbooksHelpers:  a.runbooksHelpers,
		InstallHelpers:   a.installHelpers,
		VCSHelpers:       a.vcsHelpers,
		TFClient:         a.tfClient,
	}
}
