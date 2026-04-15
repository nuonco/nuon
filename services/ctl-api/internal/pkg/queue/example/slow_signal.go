package example

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	SlowSignalType signal.SignalType = "slow-signal"
)

func init() {
	catalog.Register(SlowSignalType, func() signal.Signal {
		return &SlowSignal{}
	})
}

// SlowSignal is a test signal that blocks in Execute until the context is canceled.
type SlowSignal struct{ signal.Hooks }

var _ signal.Signal = (*SlowSignal)(nil)

func (s *SlowSignal) Validate(ctx workflow.Context) error {
	return nil
}

func (s *SlowSignal) Execute(ctx workflow.Context) error {
	// Block until context is canceled
	return workflow.Await(ctx, func() bool {
		return ctx.Err() != nil
	})
}

func (s *SlowSignal) Type() signal.SignalType {
	return SlowSignalType
}
