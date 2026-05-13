package queue

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg           *internal.Config
	V             *validator.Validate
	MetricsWriter metrics.Writer
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"context": "queue",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		cfg: params.Cfg,
		v:   params.V,
		mw:  tmw,
	}, nil
}

type Workflows struct {
	cfg *internal.Config
	v   *validator.Validate
	mw  tmetrics.Writer

	StartupHooks []func(workflow.Context, QueueWorkflowRequest) error
}

func (q *Workflows) All() []any {
	return []any{
		q.Queue,
	}
}
