package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/metrics"

	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Workflows struct {
	cfg    *internal.Config
	v      *validator.Validate
	mw     tmetrics.Writer
	logger *temporalzap.Logger
}

func (w Workflows) All() []any {
	wkflows := []any{
		w.Metrics,
	}
	return wkflows
}

// ListWorkflowFns returns the list of workflow functions for registration
func (w *Workflows) ListWorkflowFns() []any {
	return []any{
		w.Metrics,
	}
}

type WorkflowsParams struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
}

func NewWorkflows(params WorkflowsParams) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"context":   "worker",
			"namespace": defaultNamespace,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	logger, err := zap.NewProduction()
	tlogger := temporalzap.NewLogger(logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal logger: %w", err)
	}

	return &Workflows{
		cfg:    params.Cfg,
		v:      params.V,
		mw:     tmw,
		logger: tlogger,
	}, nil
}
