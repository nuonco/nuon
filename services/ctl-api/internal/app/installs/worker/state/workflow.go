package state

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
)

type Workflows struct {
	v  *validator.Validate
	mw tmetrics.Writer
}

type Params struct {
	fx.In

	V             *validator.Validate
	MetricsWriter metrics.Writer
}

func New(params Params) (*Workflows, error) {
	mw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"context": "worker",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create metrics writer: %w", err)
	}

	return &Workflows{
		v:  params.V,
		mw: mw,
	}, nil
}
