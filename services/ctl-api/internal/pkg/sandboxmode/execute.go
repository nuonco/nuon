package sandboxmode

import "go.temporal.io/sdk/workflow"

func (s *Signal) Execute(ctx workflow.Context) error {
	if cfg := s.fetchConfig(ctx); cfg != nil {
		return s.applyConfig(ctx, cfg)
	}
	return s.Signal.Execute(ctx)
}
