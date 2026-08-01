package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/triggers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	// Load the app config
	var appConfig app.AppConfig
	if res := a.db.WithContext(ctx).First(&appConfig, "id = ?", req.AppConfigID); res.Error != nil {
		return nil, fmt.Errorf("unable to load app config: %w", res.Error)
	}

	// Update status to syncing
	a.db.WithContext(ctx).Model(&appConfig).Updates(map[string]interface{}{
		"status":             app.AppConfigStatusSyncing,
		"status_description": "syncing config",
	})
	// dual-write V2 status
	syncingStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusSyncing))
	syncingStatus.StatusHumanDescription = "syncing config"
	a.db.WithContext(ctx).Model(&appConfig).Updates(map[string]any{
		"status_v2": syncingStatus,
	})

	intermediateJSON, err := appConfig.IntermediateConfig.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get intermediate config: %w", err)
	}

	var cfg config.AppConfig
	decoder := json.NewDecoder(strings.NewReader(intermediateJSON))
	decoder.UseNumber()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal intermediate config: %w", err)
	}

	// Run the DB syncer
	s := syncer.NewDBSyncer(a.db, a.helpers, a.componentHelpers, a.actionsHelpers, a.runbooksHelpers, a.installHelpers, a.vcsHelpers, a.tfClient, req.AppID, &cfg, req.AppConfigID)
	if err := s.Sync(ctx); err != nil {
		// Mark config as error
		humanErr := signal.HumanError(err)
		a.db.WithContext(ctx).Model(&appConfig).Updates(map[string]interface{}{
			"status":             app.AppConfigStatusError,
			"status_description": fmt.Sprintf("sync failed: %s", humanErr),
		})
		// dual-write V2 status
		errorStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusError))
		errorStatus.StatusHumanDescription = fmt.Sprintf("sync failed: %s", humanErr)
		a.db.WithContext(ctx).Model(&appConfig).Updates(map[string]any{
			"status_v2": errorStatus,
		})
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

	// Mark config as active with component and action IDs
	a.db.WithContext(ctx).Model(&appConfig).Updates(map[string]interface{}{
		"status":             app.AppConfigStatusActive,
		"status_description": "synced successfully",
		"component_ids":      pq.StringArray(s.GetComponentStateIds()),
		"action_ids":         pq.StringArray(s.GetActionStateIds()),
		"runbook_ids":        pq.StringArray(s.GetRunbookStateIds()),
	})
	// dual-write V2 status
	activeStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusActive))
	activeStatus.StatusHumanDescription = "synced successfully"
	if err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := triggers.Sync(ctx, tx, &cfg, appConfig.OrgID, appConfig.AppID, appConfig.ID); err != nil {
			return fmt.Errorf("unable to sync triggers: %w", err)
		}
		return tx.Model(&appConfig).Updates(map[string]interface{}{
			"status":             app.AppConfigStatusActive,
			"status_description": "synced successfully",
			"component_ids":      pq.StringArray(s.GetComponentStateIds()),
			"action_ids":         pq.StringArray(s.GetActionStateIds()),
			"status_v2":          activeStatus,
		}).Error
	}); err != nil {
		humanErr := signal.HumanError(err)
		errorStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusError))
		errorStatus.StatusHumanDescription = fmt.Sprintf("sync failed: %s", humanErr)
		a.db.WithContext(ctx).Model(&appConfig).Updates(map[string]interface{}{
			"status":             app.AppConfigStatusError,
			"status_description": fmt.Sprintf("sync failed: %s", humanErr),
			"status_v2":          errorStatus,
		})
		var syncErr configsync.SyncErr
		if errors.As(err, &syncErr) {
			return nil, temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("unable to activate app config: %s", err.Error()),
				"SYNC_VALIDATION_ERROR",
				err,
			)
		}
		return nil, fmt.Errorf("unable to activate app config: %w", err)
	}

	return &SyncAppConfigOutput{
		AppConfigID:  s.GetAppConfigID(),
		ComponentIDs: s.GetComponentStateIds(),
		ActionIDs:    s.GetActionStateIds(),
		RunbookIDs:   s.GetRunbookStateIds(),
	}, nil
}
