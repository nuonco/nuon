package example

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	CancellableSignalType signal.SignalType = "cancellable-signal"

	// CancelCallbackMarker is set by the Cancel callback and can be read via
	// the CancelCallbackInvoked query to verify the callback ran.
	CancelCallbackMarker = "cancel-callback-invoked"
)

func init() {
	catalog.Register(CancellableSignalType, func() signal.Signal {
		return &CancellableSignal{}
	})
}

// CancellableSignal blocks in Execute until cancelled and tracks whether
// its Cancel callback was invoked.
type CancellableSignal struct {
	signal.Hooks
	cancelCallbackInvoked bool
}

var _ signal.Signal = (*CancellableSignal)(nil)
var _ signal.SignalWithCancel = (*CancellableSignal)(nil)

func (c *CancellableSignal) Validate(ctx workflow.Context) error {
	return nil
}

func (c *CancellableSignal) Execute(ctx workflow.Context) error {
	// Block until context is canceled
	return workflow.Await(ctx, func() bool {
		return ctx.Err() != nil
	})
}

func (c *CancellableSignal) Cancel(ctx workflow.Context) error {
	c.cancelCallbackInvoked = true
	return nil
}

func (c *CancellableSignal) Type() signal.SignalType {
	return CancellableSignalType
}
