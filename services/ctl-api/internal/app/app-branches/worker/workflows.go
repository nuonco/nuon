package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalanalytics "github.com/nuonco/nuon/pkg/analytics/temporal"
	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/actions"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/components"
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
}

type Workflows struct {
	cfg       *internal.Config
	v         *validator.Validate
	mw        tmetrics.Writer
	evClient  teventloop.Client
	analytics temporalanalytics.Writer
	templates *cloudformation.Templates
	db        *gorm.DB

	// NOTE(sdboyer) temporary while we split up and refactor the workflows within the installs pkg
	subwfSandbox    *sandbox.Workflows
	subwfStack      *stack.Workflows
	subwfComponents *components.Workflows
	subwfActions    *actions.Workflows
}

func (w *Workflows) All() []any {
	wkflows := []any{
		w.EventLoop,
	}

	return append(wkflows, w.ListWorkflowFns())
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
	}, nil
}
