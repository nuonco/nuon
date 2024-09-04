package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Workflows struct {
	cfg      *internal.Config
	v        *validator.Validate
	acts     activities.Activities
	protos   *protos.Adapter
	mw       tmetrics.Writer
	evClient teventloop.Client
}

func NewWorkflows(v *validator.Validate,
	cfg *internal.Config,
	metricsWriter metrics.Writer,
	prt *protos.Adapter,
	evClient teventloop.Client,
) (*Workflows, error) {
	tmw, err := tmetrics.New(v,
		tmetrics.WithMetricsWriter(metricsWriter),
		tmetrics.WithTags(map[string]string{
			"namespace": signals.TemporalNamespace,
			"context":   "worker",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}
	return &Workflows{
		cfg:    cfg,
		v:      v,
		protos: prt,
		//  NOTE: this field is only used to be able to fetch activity methods
		acts:     activities.Activities{},
		mw:       tmw,
		evClient: evClient,
	}, nil
}
