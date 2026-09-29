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
	syncBranches   bool
}

type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
}

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

	s.state = &sync.State{
		Version:    "v1",
		CfgID:      s.appConfigID,
		AppID:      s.appID,
		Components: []sync.ComponentState{},
		Actions:    []sync.ActionState{},
		Runbooks:   []sync.RunbookState{},
	}

	s.prevState = &sync.State{
		Components: []sync.ComponentState{},
		Actions:    []sync.ActionState{},
		Runbooks:   []sync.RunbookState{},
	}
	s.fetchState(ctx)

	steps := s.syncSteps()

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
	}

	if s.syncBranches {
		steps = append(steps, syncStep{
			Resource: "app-branches",
			Method: func(ctx context.Context) error {
				return branches.Validate(ctx, s.db, s.cfg, s.appID)
			},
		})
	}

	steps = append(steps, []syncStep{
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
	}...)

	for _, comp := range s.cfg.Components {
		c := comp
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("component-ensure-%s", c.Name),
			Method: func(ctx context.Context) error {
				return components.EnsureComponent(ctx, s.db, s.componentHelpers, c, s.appID, s.state)
			},
		})
	}

	for _, comp := range s.cfg.Components {
		c := comp
		if len(c.Dependencies) > 0 {
			steps = append(steps, syncStep{
				Resource: fmt.Sprintf("component-deps-%s", c.Name),
				Method: func(ctx context.Context) error {
					return components.EnsureComponentDependencies(ctx, s.db, s.componentHelpers, c, s.appID)
				},
			})
		}
	}

	for _, comp := range s.cfg.Components {
		c := comp
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

	steps = append(steps, syncStep{
		Resource: "app-kubernetes-contexts",
		Method: func(ctx context.Context) error {
			return kubernetescontexts.Sync(ctx, s.db, s.cfg, s.appID, s.appConfigID)
		},
	})

	for _, action := range s.cfg.Actions {
		a := action
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("action-ensure-%s", a.Name),
			Method: func(ctx context.Context) error {
				return s.ensureAction(ctx, a)
			},
		})
	}

	for _, action := range s.cfg.Actions {
		a := action
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("action-sync-%s", a.Name),
			Method: func(ctx context.Context) error {
				return s.syncAction(ctx, a)
			},
		})
	}

	for _, runbook := range s.cfg.Runbooks {
		r := runbook
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("runbook-ensure-%s", r.Name),
			Method: func(ctx context.Context) error {
				return s.ensureRunbook(ctx, r)
			},
		})
	}

	for _, runbook := range s.cfg.Runbooks {
		r := runbook
		steps = append(steps, syncStep{
			Resource: fmt.Sprintf("runbook-sync-%s", r.Name),
			Method: func(ctx context.Context) error {
				return s.syncRunbook(ctx, r)
			},
		})
	}

	if s.syncBranches {
		// why: Branches run last: post_deploy_runbooks references runbooks by name, so the
		// runbook steps above must have created them before name resolution.
		steps = append(steps, syncStep{
			Resource: "app-branches",
			Method: func(ctx context.Context) error {
				return branches.Sync(ctx, s.db, s.appsHelpers, s.cfg, s.appID, s.state)
			},
		})
	}

	return steps
}

func (s *syncer) SyncInstall(ctx context.Context, install *config.Install) (*sync.InstallSyncResult, error) {
	return installsyncer.SyncInstall(ctx, s.db, s.installHelpers, s.appID, install)
}
