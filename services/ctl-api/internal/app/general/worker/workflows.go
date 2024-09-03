package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"

	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"

	"go.uber.org/zap"
)

type Workflows struct {
	cfg    *internal.Config
	v      *validator.Validate
	acts   activities.Activities
	protos *protos.Adapter
	mw     tmetrics.Writer
	logger *temporalzap.Logger
}

func NewWorkflows(v *validator.Validate,
	cfg *internal.Config,
	metricsWriter metrics.Writer,
	prt *protos.Adapter,
) (*Workflows, error) {

	tmw, err := tmetrics.New(v,
		tmetrics.WithMetricsWriter(metricsWriter),
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
		cfg:    cfg,
		v:      v,
		protos: prt,
		// NOTE(fd): i added these
		mw:     tmw,
		logger: tlogger,
		//  NOTE: this field is only used to be able to fetch activity methods
		acts: activities.Activities{},
	}, nil
}
