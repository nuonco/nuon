package pipeline

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

type ExecFn func(context.Context, hclog.Logger) ([]byte, error)

type CallbackFn func(context.Context, hclog.Logger, []byte) error

type Step struct {
	Name       string     `validate:"required"`
	ExecFn     ExecFn     `validate:"required" faker:"pipelineExecFn"`
	CallbackFn CallbackFn `validate:"required" faker:"pipelineCallbackFn"`
}

func (p *Pipeline) AddStep(step *Step) {
	p.Steps = append(p.Steps, step)
}
