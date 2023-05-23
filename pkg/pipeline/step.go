package pipeline

import "context"

func NoopCallbackFn(context.Context) error {
	return nil
}

type Step struct {
	Name       string      `validate:"required"`
	ExecFn     interface{} `validate:"required"`
	CallbackFn interface{} `validate:"required"`
}

func (p *Pipeline) AddStep(step *Step) {
	p.Steps = append(p.Steps, step)
}
