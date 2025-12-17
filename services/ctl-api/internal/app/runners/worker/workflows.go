package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	runner "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/kuberunner"
	teventloop "github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/temporal"
)

type Workflows struct {
	cfg      *internal.Config
	v        *validator.Validate
	mw       tmetrics.Writer
	evClient teventloop.Client
}

type WorkflowParams struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
	EvClient      teventloop.Client
}

func (w *Workflows) All() []any {
	wkflow := runner.NewWorkflow(*w.cfg)

	wkflows := []any{
		w.EventLoop,
		w.HealthCheck,
		wkflow.ProvisionRunner,
	}

	return append(wkflows, w.ListWorkflowFns()...)
}

func NewWorkflows(params WorkflowParams) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"namespace": signals.TemporalNamespace,
			"context":   "worker",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}
	return &Workflows{
		cfg:      params.Cfg,
		v:        params.V,
		mw:       tmw,
		evClient: params.EvClient,
	}, nil
}
