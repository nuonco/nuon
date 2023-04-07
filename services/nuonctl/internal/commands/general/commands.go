package general

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/repos/temporal"
)

type commands struct {
	v *validator.Validate

	Temporal temporal.Repo `validate:"required"`
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

func WithTemporalRepo(temporal temporal.Repo) commandsOption {
	return func(c *commands) error {
		c.Temporal = temporal
		return nil
	}
}
