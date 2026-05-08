package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalanalytics "github.com/nuonco/nuon/pkg/analytics/temporal"
	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/actions"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/components"
	installdelegationdns "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/dns"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/sandbox"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/stack"
	teventloop "github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

type Params struct {
	fx.In

	Cfg                 *internal.Config
	DB                  *gorm.DB `name:"psql"`
	V                   *validator.Validate
	MW                  metrics.Writer
	EVClient            teventloop.Client
	Analytics           temporalanalytics.Writer
	Templates           *cloudformation.Templates
	SandboxWorkflows    *sandbox.Workflows
	StackWorkflows      *stack.Workflows
	ComponentsWorkflows *components.Workflows
	ActionsWorkflows    *actions.Workflows
	L                   *zap.Logger
}

type Workflows struct {
	cfg       *internal.Config
	v         *validator.Validate
	mw        tmetrics.Writer
	evClient  teventloop.Client
	analytics temporalanalytics.Writer
	templates *cloudformation.Templates
	db        *gorm.DB
	l         *zap.Logger

	subwfSandbox    *sandbox.Workflows
	subwfStack      *stack.Workflows
	subwfComponents *components.Workflows
	subwfActions    *actions.Workflows
}

func (w *Workflows) All() []any {
	wkflow := installdelegationdns.NewWorkflow(*w.cfg)
	wkflows := []any{
		w.AwaitRunnerHealthy,
		w.Created,
		w.DeprovisionDNS,
		w.DeprovisionRunner,
		w.ExecuteFlow,
		w.ProvisionDNS,
		w.SyncSecrets,
		w.WorkflowApproval,
		w.Forget,
		w.PollDependencies,
		w.ProvisionRunner,
		w.ReprovisionRunner,
		w.WorkflowApproveAll,
		w.RerunFlow,
		w.Restarted,
		w.Updated,
		w.EventLoop,
		w.ActionWorkflowTriggers,
		plan.CreateActionWorkflowRunPlan,
		plan.CreateSandboxRunPlan,
		plan.CreateDeployPlan,
		plan.CreateSyncPlan,
		plan.CreateSyncSecretsPlan,
		wkflow.DeprovisionDNSDelegation,
		wkflow.ProvisionDNSDelegation,
	}

	sub := append(append(append(w.subwfSandbox.All(), w.subwfStack.All()...), w.subwfComponents.All()...), w.subwfActions.All()...)
	return append(wkflows, sub...)
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MW),
		tmetrics.WithTags(map[string]string{
			"namespace":    defaultNamespace,
			"context":      "worker",
			"actor-object": "install",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		cfg:             params.Cfg,
		v:               params.V,
		evClient:        params.EVClient,
		mw:              tmw,
		analytics:       params.Analytics,
		templates:       params.Templates,
		db:              params.DB,
		subwfSandbox:    params.SandboxWorkflows,
		subwfStack:      params.StackWorkflows,
		subwfComponents: params.ComponentsWorkflows,
		subwfActions:    params.ActionsWorkflows,
		l:               params.L,
	}, nil
}
