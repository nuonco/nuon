package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/config"
	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

	a.syncInstallsConfigRecord(ctx, req.AppID, req.AppConfigID)

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

func (a *Activities) syncInstallsConfigRecord(ctx context.Context, appID, appConfigID string) {
	var appConfig app.AppConfig
	if err := a.db.WithContext(ctx).First(&appConfig, "id = ?", appConfigID).Error; err != nil {
		return
	}

	enabled, err := a.features.OrgHasFeature(ctx, appConfig.OrgID, app.OrgFeatureAppInstallSyncing)
	if err != nil || !enabled {
		return
	}

	if appConfig.IntermediateConfig == nil || !appConfig.IntermediateConfig.IsSet() {
		return
	}

	intermediateJSON, err := appConfig.IntermediateConfig.Get(ctx)
	if err != nil || intermediateJSON == "" {
		return
	}

	var cfg config.AppConfig
	decoder := json.NewDecoder(strings.NewReader(intermediateJSON))
	decoder.UseNumber()
	if err := decoder.Decode(&cfg); err != nil {
		return
	}

	if cfg.InstallsConfig == nil {
		return
	}

	ic := cfg.InstallsConfig

	var record app.AppInstallsConfig
	record.AppID = appID
	record.Source = "config"

	if ic.ConnectedRepo != nil {
		record.VCSType = "connected"
		record.Repo = ic.ConnectedRepo.Repo
		record.Branch = ic.ConnectedRepo.Branch
		record.Directory = ic.ConnectedRepo.Directory
		if record.Directory == "" {
			record.Directory = "."
		}
	} else if ic.PublicRepo != nil {
		record.VCSType = "public"
		record.Repo = ic.PublicRepo.Repo
		record.Branch = ic.PublicRepo.Branch
		record.Directory = ic.PublicRepo.Directory
		if record.Directory == "" {
			record.Directory = "."
		}
	} else {
		return
	}

	a.db.WithContext(ctx).Create(&record)
}
