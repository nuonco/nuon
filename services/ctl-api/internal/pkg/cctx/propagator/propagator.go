package propagator

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	L *zap.Logger
}

type propagator struct{}

func New(params Params) workflow.ContextPropagator {
	return &optionalPropagator{
		l:          params.L,
		propagator: &propagator{},
	}
}
