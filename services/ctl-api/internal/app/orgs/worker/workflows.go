package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	temporalanalytics "github.com/powertoolsdev/mono/pkg/analytics/temporal"
	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"
)

type Params struct {
	fx.In

	Cfg       *internal.Config
	V         *validator.Validate
	MW        metrics.Writer
	EVClient  teventloop.Client
	Analytics temporalanalytics.Writer
	Features  *features.Features
}

type Workflows struct {
	cfg       *internal.Config
	v         *validator.Validate
	mw        tmetrics.Writer
	ev        teventloop.Client
	analytics temporalanalytics.Writer
	features  *features.Features
}

func (w *Workflows) All() []any {
	wkflows := []any{
		w.EventLoop,
	}

	return append(wkflows, w.ListWorkflowFns()...)
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MW),
		tmetrics.WithTags(map[string]string{
			"context":   "worker",
			"namespace": defaultNamespace,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		cfg:       params.Cfg,
		v:         params.V,
		mw:        tmw,
		ev:        params.EVClient,
		analytics: params.Analytics,
		features:  params.Features,
	}, nil
}
