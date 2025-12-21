package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
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

	// Deserialize the intermediate config
	intermediateJSON, err := appConfig.IntermediateConfig.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get intermediate config: %w", err)
	}

	var cfg config.AppConfig
	if err := json.Unmarshal([]byte(intermediateJSON), &cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal intermediate config: %w", err)
	}

	// Run the DB syncer
	s := syncer.New(a.db, req.AppID, req.AppBranchID, &cfg)
	result, err := s.Sync(ctx, req.AppConfigID)
	if err != nil {
		// Mark config as error
		a.db.WithContext(ctx).Model(&appConfig).Updates(map[string]interface{}{
			"status":             app.AppConfigStatusError,
			"status_description": fmt.Sprintf("sync failed: %s", err.Error()),
		})
		return nil, fmt.Errorf("unable to sync config: %w", err)
	}

	return &SyncAppConfigOutput{
		AppConfigID:  result.AppConfigID,
		ComponentIDs: result.ComponentIDs,
	}, nil
}
