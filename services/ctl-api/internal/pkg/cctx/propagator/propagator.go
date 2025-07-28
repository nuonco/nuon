package propagator

import (
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	L             *zap.Logger
	DataConverter converter.DataConverter
}

type propagator struct {
	dataConverter converter.DataConverter
}

func New(params Params) workflow.ContextPropagator {
	return &optionalPropagator{
		l: params.L,
		propagator: &propagator{
			dataConverter: params.DataConverter,
		},
	}
}
