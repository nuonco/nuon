package pipeline

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
)

// Pipeline is a type that is used to execute various commands in succession, with fail/retry logic as well as callbacks
// and others for sharing+persisting state. It's designed to power workflows such as running a terraform run which may
// involve many different steps (and outputs to s3).
//
// This is designed so that types that need to run these types of workflows can decouple the building of the steps +
// logic, from the actual execution of it.
type Pipeline struct {
	v *validator.Validate `validate:"required"`

	Steps []*Step `validate_steps:"required,gt=1"`

	Log hclog.Logger `validate:"required"`
}

type pipelineOption func(*Pipeline) error

func New(v *validator.Validate, opts ...pipelineOption) (*Pipeline, error) {
	p := &Pipeline{
		v:     v,
		Steps: make([]*Step, 0),

		Log: nil,
	}

	for idx, opt := range opts {
		if err := opt(p); err != nil {
			return nil, fmt.Errorf("unable to apply option %d: %w", idx, err)
		}
	}
	if err := p.v.Struct(p); err != nil {
		return nil, fmt.Errorf("unable to validate pipeline: %w", err)
	}

	return p, nil
}

func WithLogger(l hclog.Logger) pipelineOption {
	return func(p *Pipeline) error {
		p.Log = l
		return nil
	}
}
