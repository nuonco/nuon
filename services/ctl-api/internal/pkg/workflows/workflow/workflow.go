package workflow

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
)

type Workflows struct {
	evClient teventloop.Client
	mw       tmetrics.Writer
}

type Params struct {
	fx.In

	V             *validator.Validate
	EVClient      teventloop.Client
	MetricsWriter metrics.Writer
}

func New(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"context": "worker",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		evClient: params.EVClient,
		mw:       tmw,
	}, nil
}
