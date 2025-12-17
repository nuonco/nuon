package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	temporalanalytics "github.com/nuonco/nuon/pkg/analytics/temporal"
	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	teventloop "github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
)

type Params struct {
	fx.In

	Cfg       *internal.Config
	V         *validator.Validate
	MW        metrics.Writer
	EVClient  teventloop.Client
	Analytics temporalanalytics.Writer
	Shared    *workflows.Workflows
}

type Workflows struct {
	cfg       *internal.Config
	v         *validator.Validate
	acts      activities.Activities
	mw        tmetrics.Writer
	evClient  teventloop.Client
	analytics temporalanalytics.Writer
}

func (w *Workflows) All() []interface{} {
	wkflows := w.ListWorkflowFns()

	wkflows = append(wkflows, w.EventLoop)

	return wkflows
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MW),
		tmetrics.WithTags(map[string]string{
			"namespace": defaultNamespace,
			"context":   "worker",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		cfg:       params.Cfg,
		v:         params.V,
		evClient:  params.EVClient,
		mw:        tmw,
		analytics: params.Analytics,
	}, nil
}
