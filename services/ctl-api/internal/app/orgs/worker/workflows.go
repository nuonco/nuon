package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
)

type Params struct {
	fx.In

	Cfg      *internal.Config
	V        *validator.Validate
	MW       metrics.Writer
	EVClient teventloop.Client
}

type Workflows struct {
	cfg  *internal.Config
	v    *validator.Validate
	acts activities.Activities
	mw   tmetrics.Writer
	ev   teventloop.Client
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
		cfg: params.Cfg,
		v:   params.V,
		mw:  tmw,
		ev:  params.EVClient,
		//  NOTE: this field is only used to be able to fetch activity methods
		acts: activities.Activities{},
	}, nil
}
