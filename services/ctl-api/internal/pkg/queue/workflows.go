package queue

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal"
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

	StartupHooks []func(workflow.Context, QueueWorkflowRequest) error
}

func (q *Workflows) All() []any {
	return []any{
		q.Queue,
	}
}
