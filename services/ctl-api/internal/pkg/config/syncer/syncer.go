package syncer

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

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

// NewDBSyncer creates a database-backed syncer for use in Temporal workflows.
// The context must contain org and account information before calling Sync().
func NewDBSyncer(db *gorm.DB, appID string, cfg *config.AppConfig, appConfigID string) sync.Syncer {
	return &syncer{
		db:          db,
		cfg:         cfg,
		appID:       appID,
		appConfigID: appConfigID,
		state:       nil, // will be populated by fetchState()
		prevState:   nil,
	}
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

func (s *syncer) syncSteps() []syncStep {
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

// NOTE: syncComponent() and finish() methods are defined in components.go and app_config.go respectively
