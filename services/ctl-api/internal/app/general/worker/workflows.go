package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/metrics"

	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Workflows struct {
	cfg    *internal.Config
	v      *validator.Validate
	protos *protos.Adapter
	mw     tmetrics.Writer
	logger *temporalzap.Logger
	ev     teventloop.Client
}

func (w Workflows) All() []any {
	wkflows := []any{
		w.EventLoop,
		w.Metrics,
		w.Promotion,
		w.RestartOrgRunners,
		w.RestartOrgEventLoops,
	}
	return append(wkflows)
}

type WorkflowsParams struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
	Prt           *protos.Adapter
	EVClient      teventloop.Client
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
		protos: params.Prt,
		ev:     params.EVClient,
		mw:     tmw,
		logger: tlogger,
	}, nil
}
