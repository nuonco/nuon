package signal

import "go.temporal.io/sdk/workflow"

// SignalWithUpdateHandlers is an optional interface that signals can implement
// to register custom Temporal update handlers on the handler workflow.
// Called after initializeState() but before the handler is marked ready,
// so handlers are available for the entire lifecycle of the signal.
type SignalWithUpdateHandlers interface {
	Signal

	RegisterUpdateHandlers(ctx workflow.Context) error
}

// RegisterUpdateHandlers checks if the signal implements SignalWithUpdateHandlers
// and calls RegisterUpdateHandlers if so.
func RegisterUpdateHandlers(sig Signal, ctx workflow.Context) error {
	if su, ok := sig.(SignalWithUpdateHandlers); ok {
		return su.RegisterUpdateHandlers(ctx)
	}
	return nil
}
