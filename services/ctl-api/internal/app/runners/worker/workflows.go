package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	runner "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/kuberunner"
)

type Workflows struct {
	cfg *internal.Config
	v   *validator.Validate
	mw  tmetrics.Writer
}

type WorkflowParams struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
}

func (w *Workflows) All() []any {
	wkflow := runner.NewWorkflow(*w.cfg)

	return []any{
		w.CronShutdownVM,
		wkflow.ProvisionRunner,
		wkflow.DeprovisionRunner,
	}
}

func NewWorkflows(params WorkflowParams) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"namespace": "runners",
			"context":   "worker",
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
