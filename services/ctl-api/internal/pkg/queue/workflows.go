package queue

import (
	"go.uber.org/fx"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	V   *validator.Validate
}

func NewWorkflows(params Params) (*Workflows, error) {
	return &Workflows{
		cfg: params.Cfg,
		v:   params.V,
	}, nil
}

type Workflows struct {
	cfg *internal.Config
	v   *validator.Validate
}

func (q *Workflows) All() []any {
	return []any{
		q.Queue,
	}
}
