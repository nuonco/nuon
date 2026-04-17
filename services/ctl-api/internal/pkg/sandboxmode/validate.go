package sandboxmode

import "go.temporal.io/sdk/workflow"

func (s *Signal) Validate(ctx workflow.Context) error {
	cfg := s.fetchConfig(ctx)
	if cfg != nil {
		return s.applyConfig(ctx, cfg)
	}
	return s.Signal.Validate(ctx)
}
