package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
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
	wkflows := []any{
		w.EventLoop,
		w.HealthCheck,
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
