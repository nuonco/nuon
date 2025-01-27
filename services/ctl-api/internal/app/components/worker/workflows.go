package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Workflows struct {
	cfg      *internal.Config
	v        *validator.Validate
	protos   *protos.Adapter
	mw       tmetrics.Writer
	evClient teventloop.Client
}

func (w *Workflows) All() []any {
	fns := w.ListWorkflowFns()
	fns = append(fns, w.EventLoop)
	return fns
}

type WorkflowsParams struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
	Prt           *protos.Adapter
	EvClient      teventloop.Client
}

func NewWorkflows(params WorkflowsParams) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"namespace": defaultNamespace,
			"context":   "worker",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}
	return &Workflows{
		cfg:      params.Cfg,
		v:        params.V,
		protos:   params.Prt,
		mw:       tmw,
		evClient: params.EvClient,
	}, nil
}
