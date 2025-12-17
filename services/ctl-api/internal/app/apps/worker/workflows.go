package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	temporalanalytics "github.com/nuonco/nuon/pkg/analytics/temporal"
	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/ecrrepository"
)

type Workflows struct {
	cfg       *internal.Config
	v         *validator.Validate
	acts      activities.Activities
	mw        tmetrics.Writer
	analytics temporalanalytics.Writer
}

func (w *Workflows) All() []any {
	var wkflow ecrrepository.Wkflow
	wkflows := []any{
		w.EventLoop,
		wkflow.ProvisionECRRepository,
		wkflow.DeprovisionECRRepository,
	}

	return append(wkflows, w.ListWorkflowFns()...)
}

type Params struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
	Analytics     temporalanalytics.Writer
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter), tmetrics.WithTags(map[string]string{
			"namespace": defaultNamespace,
			"context":   "worker",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		cfg:       params.Cfg,
		v:         params.V,
		mw:        tmw,
		analytics: params.Analytics,
	}, nil
}
