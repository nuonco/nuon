package activities

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer"
)

type SyncCustomAppConfigOutput struct {
	AppConfigID  string   `json:"app_config_id"`
	ComponentIDs []string `json:"component_ids"`
	ActionIDs    []string `json:"action_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) syncCustomAppConfig(ctx context.Context, onboardingID string) (*SyncCustomAppConfigOutput, error) {
	// Fetch onboarding — it has AppID, AppConfig, and everything we need
	var onboarding app.Onboarding
	if err := a.db.WithContext(ctx).First(&onboarding, "id = ?", onboardingID).Error; err != nil {
		return nil, fmt.Errorf("unable to get onboarding: %w", err)
	}

	if onboarding.AppConfig == nil {
		return nil, fmt.Errorf("onboarding has no app config built")
	}
	if onboarding.AppID == nil || *onboarding.AppID == "" {
		return nil, fmt.Errorf("onboarding has no app_id set")
	}

	cfg := &onboarding.AppConfig.AppConfig
	appID := *onboarding.AppID

	// Create AppConfig record for the syncer to link child records to
	appConfig := &app.AppConfig{
		AppID:             appID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "pending sync",
	}

	if err := a.db.WithContext(ctx).Create(appConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to create app config: %w", err)
	}

	// Update status to syncing
	a.db.WithContext(ctx).Model(appConfig).Updates(map[string]interface{}{
		"status":             app.AppConfigStatusSyncing,
		"status_description": "syncing config",
	})

	// Run the DB syncer to create components, sandbox, runner, etc.
	s := syncer.NewDBSyncer(a.db, appID, cfg, appConfig.ID)
	if err := s.Sync(ctx); err != nil {
		a.db.WithContext(ctx).Model(appConfig).Updates(map[string]interface{}{
			"status":             app.AppConfigStatusError,
			"status_description": fmt.Sprintf("sync failed: %s", err.Error()),
		})
		return nil, fmt.Errorf("unable to sync config: %w", err)
	}

	// Mark config as active with component and action IDs
	a.db.WithContext(ctx).Model(appConfig).Updates(map[string]interface{}{
		"status":             app.AppConfigStatusActive,
		"status_description": "synced successfully",
		"component_ids":      pq.StringArray(s.GetComponentStateIds()),
		"action_ids":         pq.StringArray(s.GetActionStateIds()),
	})

	return &SyncCustomAppConfigOutput{
		AppConfigID:  appConfig.ID,
		ComponentIDs: s.GetComponentStateIds(),
		ActionIDs:    s.GetActionStateIds(),
	}, nil
}
