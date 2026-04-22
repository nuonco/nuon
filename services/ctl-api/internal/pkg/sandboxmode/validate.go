package sandboxmode

import (
	"errors"

	"go.temporal.io/sdk/workflow"
)

func (s *Signal) Validate(ctx workflow.Context) error {
	if cfg := s.fetchConfig(ctx); cfg != nil && cfg.ValidateError != "" {
		return errors.New(cfg.ValidateError)
	}

	return s.Signal.Validate(ctx)
}
