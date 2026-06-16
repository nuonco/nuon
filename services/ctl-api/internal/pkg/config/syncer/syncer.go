package syncer

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/appconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/breakglass"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/components"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/inputs"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/policies"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/runner"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/sandbox"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/secrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/stack"
)

// syncer implements sync.Syncer using direct database access.
// This implementation is used by workflows within ctl-api to sync configs
// without going through HTTP endpoints.
type syncer struct {
	db               *gorm.DB
	cfg              *config.AppConfig
	componentHelpers *componenthelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	runbooksHelpers  *runbookshelpers.Helpers

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
func NewDBSyncer(db *gorm.DB, componentHelpers *componenthelpers.Helpers, actionsHelpers *actionshelpers.Helpers, runbooksHelpers *runbookshelpers.Helpers, appID string, cfg *config.AppConfig, appConfigID string) sync.Syncer {
	return &syncer{
		db:               db,
		cfg:              cfg,
		componentHelpers: componentHelpers,
		actionsHelpers:   actionsHelpers,
		runbooksHelpers:  runbooksHelpers,
		appID:            appID,
		appConfigID:      appConfigID,
		state:            nil, // will be populated by fetchState()
		prevState:        nil,
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

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return sync.SyncInternalErr{
			Description: "missing org context",
			Err:         err,
		}
	}
	s.orgID = orgID

	// Initialize state
	s.state = &sync.State{
		Version:    "v1",
		CfgID:      s.appConfigID,
		AppID:      s.appID,
		Components: []sync.ComponentState{},
		Actions:    []sync.ActionState{},
		Runbooks:   []sync.RunbookState{},
	}

	// Initialize prevState for orphaned resource tracking
	s.prevState = &sync.State{
		Components: []sync.ComponentState{},
		Actions:    []sync.ActionState{},
		Runbooks:   []sync.RunbookState{},
	}

	// Build sync steps
	steps := s.syncSteps()

	// Execute sync steps
	for _, step := range steps {
		if err := step.Method(ctx); err != nil {
			return err
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
			Resource: "app-config-metadata",
			Method: func(ctx context.Context) error {
				return appconfig.Sync(ctx, s.db, s.cfg, s.appConfigID)
			},
		},
		{
			Resource: "app-inputs",
			Method: func(ctx context.Context) error {
				return inputs.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID, s.orgID, s.state)
			},
		},
		{
			Resource: "app-sandbox",
			Method: func(ctx context.Context) error {
				return sandbox.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID, s.state)
			},
		},
		{
			Resource: "app-runner",
			Method: func(ctx context.Context) error {
				return runner.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID, s.state)
			},
		},
		{
			Resource: "app-permissions",
			Method: func(ctx context.Context) error {
				return permissions.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
			},
		},
		{
			Resource: "app-policies",
			Method: func(ctx context.Context) error {
				return policies.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
			},
		},
		{
			Resource: "app-secrets",
			Method: func(ctx context.Context) error {
				return secrets.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
			},
		},
		{
			Resource: "app-break-glass",
			Method: func(ctx context.Context) error {
				return breakglass.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
			},
		},
		{
			Resource: "app-cloudformation-stack",
			Method: func(ctx context.Context) error {
				return stack.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
			},
		},
	}

	// Ensure all components exist (with full initialization: queue, dependencies, install components)
	for _, comp := range s.cfg.Components {
		c := comp // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("component-ensure-%s", c.Name),
			Method: func(ctx context.Context) error {
				return components.EnsureComponent(ctx, s.db, s.componentHelpers, c, s.appID)
			},
		})
	}

	// Resolve component dependencies (after all components exist)
	for _, comp := range s.cfg.Components {
		c := comp // Capture loop variable
		if len(c.Dependencies) > 0 {
			steps = append(steps, syncStep{
				Resource: fmt.Sprintf("component-deps-%s", c.Name),
				Method: func(ctx context.Context) error {
					return components.EnsureComponentDependencies(ctx, s.db, s.componentHelpers, c, s.appID)
				},
			})
		}
	}

	// Sync component configurations
	for _, comp := range s.cfg.Components {
		c := comp // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("component-sync-%s", c.Name),
			Method: func(ctx context.Context) error {
				return components.SyncComponent(ctx, s.db, s.componentHelpers, c, s.appID, s.appConfigID, s.state)
			},
		})
	}

	// Ensure all actions exist (with full initialization: install action workflows)
	for _, action := range s.cfg.Actions {
		a := action // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("action-ensure-%s", a.Name),
			Method: func(ctx context.Context) error {
				return s.ensureAction(ctx, a)
			},
		})
	}

	// Sync action configurations
	for _, action := range s.cfg.Actions {
		a := action // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("action-sync-%s", a.Name),
			Method: func(ctx context.Context) error {
				return s.syncAction(ctx, a)
			},
		})
	}

	// Ensure all runbooks exist (with full initialization: install runbooks)
	for _, runbook := range s.cfg.Runbooks {
		r := runbook // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("runbook-ensure-%s", r.Name),
			Method: func(ctx context.Context) error {
				return s.ensureRunbook(ctx, r)
			},
		})
	}

	// Sync runbook configurations
	for _, runbook := range s.cfg.Runbooks {
		r := runbook // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("runbook-sync-%s", r.Name),
			Method: func(ctx context.Context) error {
				return s.syncRunbook(ctx, r)
			},
		})
	}

	return steps
}

// NOTE: syncComponent() and finish() methods are defined in components.go and app_config.go respectively
