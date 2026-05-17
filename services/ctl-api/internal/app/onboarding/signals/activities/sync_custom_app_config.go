package activities

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
	pendingStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusPending))
	pendingStatus.StatusHumanDescription = "pending sync"

	appConfig := &app.AppConfig{
		AppID:             appID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "pending sync",
		StatusV2:          pendingStatus,
	}

	if err := a.db.WithContext(ctx).Create(appConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to create app config: %w", err)
	}

	// Update status to syncing
	a.db.WithContext(ctx).Model(appConfig).Updates(map[string]interface{}{
		"status":             app.AppConfigStatusSyncing,
		"status_description": "syncing config",
	})
	// dual-write V2 status
	syncingStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusSyncing))
	syncingStatus.StatusHumanDescription = "syncing config"
	a.db.WithContext(ctx).Model(appConfig).Updates(map[string]any{
		"status_v2": syncingStatus,
	})

	// Run the DB syncer to create components, sandbox, runner, etc.
	s := syncer.NewDBSyncer(a.db, a.componentHelpers, a.actionsHelpers, a.runbooksHelpers, appID, cfg, appConfig.ID)
	if err := s.Sync(ctx); err != nil {
		humanErr := signal.HumanError(err)
		a.db.WithContext(ctx).Model(appConfig).Updates(map[string]interface{}{
			"status":             app.AppConfigStatusError,
			"status_description": fmt.Sprintf("sync failed: %s", humanErr),
		})
		// dual-write V2 status
		errorStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusError))
		errorStatus.StatusHumanDescription = fmt.Sprintf("sync failed: %s", humanErr)
		a.db.WithContext(ctx).Model(appConfig).Updates(map[string]any{
			"status_v2": errorStatus,
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
	// dual-write V2 status
	activeStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusActive))
	activeStatus.StatusHumanDescription = "synced successfully"
	a.db.WithContext(ctx).Model(appConfig).Updates(map[string]any{
		"status_v2": activeStatus,
	})

	return &SyncCustomAppConfigOutput{
		AppConfigID:  appConfig.ID,
		ComponentIDs: s.GetComponentStateIds(),
		ActionIDs:    s.GetActionStateIds(),
	}, nil
}
