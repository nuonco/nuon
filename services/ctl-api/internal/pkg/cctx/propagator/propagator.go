package propagator

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
}

type propagator struct{}

func New() workflow.ContextPropagator {
	return &propagator{}
}
