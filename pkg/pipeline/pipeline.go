package pipeline

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"go.uber.org/zap"
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

	Log *zap.Logger `validate:"required"`
	UI  terminal.UI `validate:"required"`
}

type pipelineOption func(*Pipeline) error

func New(v *validator.Validate, opts ...pipelineOption) (*Pipeline, error) {
	p := &Pipeline{
		v:     v,
		Steps: make([]*Step, 0),

		Log: nil,
		UI:  nil,
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

func WithUI(ui terminal.UI) pipelineOption {
	return func(p *Pipeline) error {
		p.UI = ui
		return nil
	}
}

func WithLogger(l *zap.Logger) pipelineOption {
	return func(p *Pipeline) error {
		p.Log = l
		return nil
	}
}
