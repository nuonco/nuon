package workflow

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
)

type Workflows struct {
	mw tmetrics.Writer
	v  *validator.Validate
}

type Params struct {
	fx.In

	V             *validator.Validate
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
		mw: tmw,
		v:  params.V,
	}, nil
}
