package syncer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// start creates a new AppConfig record with status pending.
func (s *syncer) start(ctx context.Context) error {
	cfg := app.AppConfig{
		OrgID:             s.orgID,
		AppID:             s.appID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "sync pending",
		Readme:            s.cfg.Readme,
		CLIVersion:        "", // No CLI version for workflow-based sync
	}

	res := s.db.WithContext(ctx).Create(&cfg)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app config",
			Err:         fmt.Errorf("unable to create app config: %w", res.Error),
		}
	}

	s.appConfigID = cfg.ID
	s.state.CfgID = cfg.ID
	return nil
}

// finish updates the AppConfig status to success and saves the final state.
func (s *syncer) finish(ctx context.Context) error {
	// Marshal state to JSON
	stateJSON, err := json.Marshal(s.state)
	if err != nil {
		return fmt.Errorf("unable to marshal state: %w", err)
	}

	// Build component IDs list
	compIDs := make([]string, 0)
	for _, comp := range s.state.Components {
		compIDs = append(compIDs, comp.ID)
	}

	// Find the app config
	var cfg app.AppConfig
	if err := s.db.WithContext(ctx).
		Where("id = ?", s.appConfigID).
		First(&cfg).Error; err != nil {
		return fmt.Errorf("unable to find app config: %w", err)
	}

	// Update the app config
	updates := app.AppConfig{
		Status:            app.AppConfigStatusActive,
		StatusDescription: "successfully synced config",
		State:             string(stateJSON),
		ComponentIDs:      compIDs,
	}

	res := s.db.WithContext(ctx).
		Model(&cfg).
		Updates(updates)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to update app config status",
			Err:         res.Error,
		}
	}

	return nil
}

// fetchState loads the most recent successful AppConfig state.
func (s *syncer) fetchState(ctx context.Context) error {
	var prevConfig app.AppConfig
	res := s.db.WithContext(ctx).
		Where("app_id = ? AND status = ?", s.appID, app.AppConfigStatusActive).
		Order("created_at DESC").
		First(&prevConfig)

	if res.Error != nil {
		// No previous state is OK
		return nil
	}

	if prevConfig.State != "" {
		if err := json.Unmarshal([]byte(prevConfig.State), s.prevState); err != nil {
			return fmt.Errorf("unable to unmarshal previous state: %w", err)
		}
	}

	return nil
}
