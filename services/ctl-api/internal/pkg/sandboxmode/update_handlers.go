package sandboxmode

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if su, ok := s.Signal.(signal.SignalWithUpdateHandlers); ok {
		return su.RegisterUpdateHandlers(ctx)
	}

	return nil
}
