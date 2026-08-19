package syncer

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/appconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/branches"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/breakglass"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/components"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/inputs"
	installsyncer "github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/installs"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/kubernetescontexts"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/operationroles"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/policies"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/runner"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/sandbox"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/secrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/stack"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
)

// syncer implements sync.Syncer using direct database access.
// This implementation is used by workflows within ctl-api to sync configs
// without going through HTTP endpoints.
type syncer struct {
	db               *gorm.DB
	cfg              *config.AppConfig
	appsHelpers      *appshelpers.Helpers
	componentHelpers *componenthelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	runbooksHelpers  *runbookshelpers.Helpers
	installHelpers   *installhelpers.Helpers
	vcsHelpers       *vcshelpers.Helpers
	tfClient         terraform.Client

	appID       string
	appConfigID string
	orgID       string

	state     *sync.State
	prevState *sync.State

	dispatchBuilds bool
}

// Params defines the dependencies required by the syncer.
// This follows the FX dependency injection pattern used in ctl-api.
type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
}

// NewDBSyncer creates a database-backed syncer for use in Temporal workflows.
// The context must contain org and account information before calling Sync().
func NewDBSyncer(db *gorm.DB, appsHelpers *appshelpers.Helpers, componentHelpers *componenthelpers.Helpers, actionsHelpers *actionshelpers.Helpers, runbooksHelpers *runbookshelpers.Helpers, installHelpers *installhelpers.Helpers, vcsHelpers *vcshelpers.Helpers, tfClient terraform.Client, appID string, cfg *config.AppConfig, appConfigID string, opts ...Option) sync.Syncer {
	s := &syncer{
		db:               db,
		cfg:              cfg,
		appsHelpers:      appsHelpers,
		componentHelpers: componentHelpers,
		actionsHelpers:   actionsHelpers,
		runbooksHelpers:  runbooksHelpers,
		installHelpers:   installHelpers,
		vcsHelpers:       vcsHelpers,
		tfClient:         tfClient,
		appID:            appID,
		appConfigID:      appConfigID,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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
	if err := s.validateFeatureCompatibility(ctx); err != nil {
		return err
	}

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
	s.fetchState(ctx)

	// Build sync steps
	steps := s.syncSteps()

	// Execute sync steps
	for _, step := range steps {
		if err := step.Method(ctx); err != nil {
			return err
		}
	}

	return s.persistState(ctx)
}

func (s *syncer) validateFeatureCompatibility(ctx context.Context) error {
	var org app.Org
	res := s.db.WithContext(ctx).
		Select("id", "features").
		Where(&app.Org{ID: s.orgID}).
		First(&org)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to check org feature compatibility",
			Err:         res.Error,
		}
	}
	if s.cfg.Triggers != nil && len(s.cfg.Triggers.Rules) != 0 && !org.Features[string(app.OrgFeatureTriggers)] {
		return sync.SyncErr{Resource: "triggers", Description: "the triggers feature is not enabled for this organization"}
	}
	if s.cfg.Sandbox != nil && s.cfg.Sandbox.Type == config.AppSandboxTypePulumi && !org.Features[string(app.OrgFeaturePulumiSandbox)] {
		return sync.SyncErr{Resource: "app-sandbox", Description: "pulumi sandboxes are not enabled for this organization"}
	}

	return sync.RejectDockerBuildComponentsForFeature(s.cfg)
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
			// Validate branches early even though they are written last, so a bad
			// branch block fails before components sync and builds are dispatched.
			Resource: "app-branches",
			Method: func(ctx context.Context) error {
				return branches.Validate(ctx, s.db, s.cfg, s.appID)
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
				return sandbox.Sync(ctx, s.db, s.vcsHelpers, s.cfg, s.appID, s.appConfigID, s.state)
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
				return permissions.Sync(ctx, s.db, s.installHelpers, s.cfg, s.appID, s.appConfigID)
			},
		},
		{
			Resource: "app-operations-roles",
			Method: func(ctx context.Context) error {
				return operationroles.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
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
				return stack.Sync(ctx, s.db, s.appsHelpers, s.cfg, s.appID, s.appConfigID)
			},
		},
	}

	// Ensure all components exist (with full initialization: queue, dependencies, install components)
	for _, comp := range s.cfg.Components {
		c := comp // Capture loop variable
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("component-ensure-%s", c.Name),
			Method: func(ctx context.Context) error {
				return components.EnsureComponent(ctx, s.db, s.componentHelpers, c, s.appID, s.state)
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
				return components.SyncComponent(ctx, components.SyncComponentParams{
					DB:             s.db,
					Helpers:        s.componentHelpers,
					VCSHelper:      s.vcsHelpers,
					TFClient:       s.tfClient,
					Component:      c,
					AppID:          s.appID,
					AppConfigID:    s.appConfigID,
					State:          s.state,
					DispatchBuilds: s.dispatchBuilds,
				})
			},
		})
	}

	// Sync kubernetes contexts after components: each context resolves its
	// source-component name to an ID, which only exists once components are synced.
	steps = append(steps, syncStep{
		Resource: "app-kubernetes-contexts",
		Method: func(ctx context.Context) error {
			return kubernetescontexts.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
		},
	})

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

	// Branches run last: post_deploy_runbooks references runbooks by name, so the
	// runbook steps above must have created them before name resolution.
	steps = append(steps, syncStep{
		Resource: "app-branches",
		Method: func(ctx context.Context) error {
			return branches.Sync(ctx, s.db, s.appsHelpers, s.cfg, s.appID, s.state)
		},
	})

	return steps
}

func (s *syncer) SyncInstall(ctx context.Context, install *config.Install) (*sync.InstallSyncResult, error) {
	return installsyncer.SyncInstall(ctx, s.db, s.installHelpers, s.appID, install)
}

// NOTE: syncComponent() and finish() methods are defined in components.go and app_config.go respectively
