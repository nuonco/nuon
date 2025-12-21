package syncer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SyncResult struct {
	AppConfigID  string
	ComponentIDs []string
}

type Syncer struct {
	db          *gorm.DB
	appID       string
	appBranchID string
	cfg         *config.AppConfig
	appConfigID string

	// state tracked during sync
	componentIDs []string
}

func New(db *gorm.DB, appID string, appBranchID string, cfg *config.AppConfig) *Syncer {
	return &Syncer{
		db:           db,
		appID:        appID,
		appBranchID:  appBranchID,
		cfg:          cfg,
		componentIDs: make([]string, 0),
	}
}

type syncStep struct {
	Resource string
	Method   func(context.Context) error
}

func (s *Syncer) Sync(ctx context.Context, appConfigID string) (*SyncResult, error) {
	s.appConfigID = appConfigID

	if s.cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	steps := s.syncSteps()
	for _, step := range steps {
		if err := step.Method(ctx); err != nil {
			return nil, fmt.Errorf("sync step %s failed: %w", step.Resource, err)
		}
	}

	// Update app config with component IDs and active status
	if err := s.finish(ctx); err != nil {
		return nil, fmt.Errorf("unable to finish sync: %w", err)
	}

	return &SyncResult{
		AppConfigID:  s.appConfigID,
		ComponentIDs: s.componentIDs,
	}, nil
}

func (s *Syncer) syncSteps() []syncStep {
	steps := []syncStep{}

	// Sync components - ensure they exist and create/update configs
	for _, comp := range s.cfg.Components {
		comp := comp
		resourceName := fmt.Sprintf("component-%s", comp.Name)
		steps = append(steps, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				return s.syncComponent(ctx, comp)
			},
		})
	}

	// TODO: Add remaining sync steps as needed:
	// - inputs
	// - sandbox
	// - runner
	// - permissions
	// - policies
	// - secrets
	// - break-glass
	// - stack
	// - actions

	return steps
}

func (s *Syncer) syncComponent(ctx context.Context, comp *config.Component) error {
	// Find or create the component
	var existing app.Component
	err := s.db.WithContext(ctx).
		Where("app_id = ? AND name = ?", s.appID, comp.Name).
		First(&existing).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to find component %s: %w", comp.Name, err)
	}

	var componentID string
	if err == gorm.ErrRecordNotFound {
		// Create the component
		newComp := app.Component{
			AppID:             s.appID,
			Name:              comp.Name,
			VarName:           comp.VarName,
			Status:            app.ComponentStatusActive,
			StatusDescription: "synced from config",
		}
		if res := s.db.WithContext(ctx).Create(&newComp); res.Error != nil {
			return fmt.Errorf("unable to create component %s: %w", comp.Name, res.Error)
		}
		componentID = newComp.ID
	} else {
		componentID = existing.ID
		// Update component fields
		if res := s.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"var_name": comp.VarName,
		}); res.Error != nil {
			return fmt.Errorf("unable to update component %s: %w", comp.Name, res.Error)
		}
	}

	// Create component config connection for this app config
	connection := app.ComponentConfigConnection{
		AppConfigID: s.appConfigID,
		ComponentID: componentID,
	}
	if res := s.db.WithContext(ctx).Create(&connection); res.Error != nil {
		return fmt.Errorf("unable to create component config connection for %s: %w", comp.Name, res.Error)
	}

	s.componentIDs = append(s.componentIDs, componentID)
	return nil
}

func (s *Syncer) finish(ctx context.Context) error {
	stateJSON, err := json.Marshal(map[string]interface{}{
		"version":       "v1",
		"app_id":        s.appID,
		"config_id":     s.appConfigID,
		"component_ids": s.componentIDs,
	})
	if err != nil {
		return fmt.Errorf("unable to marshal state: %w", err)
	}

	updates := map[string]interface{}{
		"status":             app.AppConfigStatusActive,
		"status_description": "successfully synced config",
		"state":              string(stateJSON),
		"component_ids":      pq.StringArray(s.componentIDs),
		"app_branch_id":      generics.NewNullString(s.appBranchID),
	}

	if res := s.db.WithContext(ctx).
		Model(&app.AppConfig{}).
		Where("id = ?", s.appConfigID).
		Updates(updates); res.Error != nil {
		return fmt.Errorf("unable to update app config: %w", res.Error)
	}

	return nil
}
