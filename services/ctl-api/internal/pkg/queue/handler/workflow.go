package handler

import (
	"go.uber.org/fx"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg           *internal.Config
	V             *validator.Validate
	MetricsWriter metrics.Writer
}

func NewWorkflows(params Params) (*Workflows, error) {
	return &Workflows{
		cfg: params.Cfg,
		v:   params.V,
		mw:  params.MetricsWriter,
	}, nil
}

type Workflows struct {
	cfg *internal.Config
	v   *validator.Validate
	mw  metrics.Writer
}

func (q *Workflows) All() []any {
	return []any{
		q.Handler,
	}
}
