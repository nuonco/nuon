package pipeline

import (
	"context"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"go.uber.org/zap"
)

// ExecFn is a function used to execute a step
type ExecFn func(context.Context, *zap.Logger, terminal.UI) ([]byte, error)

// CallbackFn is a function used to send the outputs of an exec, as a callback
type CallbackFn func(context.Context, *zap.Logger, terminal.UI, []byte) error

type Step struct {
	Name       string     `validate:"required"`
	ExecFn     ExecFn     `validate:"required" faker:"pipelineExecFn"`
	CallbackFn CallbackFn `validate:"required" faker:"pipelineCallbackFn"`
}

func (p *Pipeline) AddStep(step *Step) {
	p.Steps = append(p.Steps, step)
}
