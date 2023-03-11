package installs

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/repos/executors"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/repos/temporal"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/repos/workflows"
)

type commands struct {
	v *validator.Validate

	Workflows workflows.Repo `validate:"required"`
	Temporal  temporal.Repo  `validate:"required"`
	Executors executors.Repo `validate:"required"`
}

// New returns a default commands with the default orgcontext getter
func New(v *validator.Validate, opts ...commandsOption) (*commands, error) {
	r := &commands{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	return r, nil
}

type commandsOption func(*commands) error

func WithWorkflowsRepo(repo workflows.Repo) commandsOption {
	return func(c *commands) error {
		c.Workflows = repo
		return nil
	}
}

func WithTemporalRepo(temporal temporal.Repo) commandsOption {
	return func(c *commands) error {
		c.Temporal = temporal
		return nil
	}
}

func WithExecutorsRepo(executors executors.Repo) commandsOption {
	return func(c *commands) error {
		c.Executors = executors
		return nil
	}
}
