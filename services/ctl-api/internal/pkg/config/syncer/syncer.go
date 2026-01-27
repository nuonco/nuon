package syncer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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

// syncer implements sync.Syncer using direct database access.
// This implementation is used by workflows within ctl-api to sync configs
// without going through HTTP endpoints.
type syncer struct {
	db  *gorm.DB
	cfg *config.AppConfig

	appID       string
	appConfigID string
	orgID       string

	state     *sync.State
	prevState *sync.State

	cmpBuildsScheduled []string
}

// Params defines the dependencies required by the syncer.
// This follows the FX dependency injection pattern used in ctl-api.
type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
}

// New creates a new database-based syncer that directly accesses the database.
// This is used by Temporal workflows within ctl-api.
//
// The context must contain org and account information set via:
//   - cctx.SetOrgContext()
//   - cctx.SetAccountContext()
//
// Parameters:
//   - p: FX Params struct containing gorm.DB dependency
//   - appID: ID of the app to sync
//   - cfg: parsed app configuration to sync
//
// Returns a sync.Syncer interface that can be used to perform the sync operation.
func New(p Params, appID string, cfg *config.AppConfig) sync.Syncer {
	return &syncer{
		db:    p.DB,
		cfg:   cfg,
		appID: appID,
		state: &sync.State{
			Version: sync.DefaultStateVersion,
			AppID:   appID,
		},
		prevState:          &sync.State{},
		cmpBuildsScheduled: make([]string, 0),
	}
}

// Sync implements sync.Syncer
func (s *syncer) Sync(ctx context.Context) error {
	s.cmpBuildsScheduled = make([]string, 0)

	if s.cfg == nil {
		return sync.SyncInternalErr{
			Description: "nil config",
			Err:         fmt.Errorf("config is nil"),
		}
	}

	// Extract org ID from context
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return sync.SyncInternalErr{
			Description: "missing org context",
			Err:         err,
		}
	}
	s.orgID = org.ID

	// Fetch previous state
	if err := s.fetchState(ctx); err != nil {
		return sync.SyncInternalErr{
			Description: "unable to fetch state",
			Err:         err,
		}
	}

	// Create app config
	if err := s.start(ctx); err != nil {
		return sync.SyncInternalErr{
			Description: "unable to start sync",
			Err:         err,
		}
	}

	// Build sync steps
	steps := s.syncSteps()

	// Execute sync steps
	for _, step := range steps {
		if err := step.Method(ctx); err != nil {
			return err
		}
	}

	// Mark config as complete
	if err := s.finish(ctx); err != nil {
		return sync.SyncInternalErr{
			Description: "unable to finish sync",
			Err:         err,
		}
	}

	return nil
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
	steps := []syncStep{
		{
			Resource: "app",
			Method:   s.syncApp,
		},
		{
			Resource: "app-inputs",
			Method:   s.syncAppInput,
		},
		{
			Resource: "app-sandbox",
			Method:   s.syncAppSandbox,
		},
		{
			Resource: "app-runner",
			Method:   s.syncAppRunner,
		},
		{
			Resource: "app-permissions",
			Method:   s.syncAppPermissions,
		},
		{
			Resource: "app-policies",
			Method:   s.syncAppPolicies,
		},
		{
			Resource: "app-secrets",
			Method:   s.syncAppSecrets,
		},
		{
			Resource: "app-break-glass",
			Method:   s.syncAppBreakGlass,
		},
		{
			Resource: "app-cloudformation-stack",
			Method:   s.syncAppCloudFormationStack,
		},
	}

	// Ensure all components exist
	for _, comp := range s.cfg.Components {
		c := comp // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("component-%s", c.Name),
			Method: func(ctx context.Context) error {
				return s.ensureComponent(ctx, c)
			},
		})
	}

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
