package signal

import "go.temporal.io/sdk/workflow"

// SignalWithInit is an optional interface that signals can implement to
// perform initialization after params have been applied but before
// validate/execute are called.
type SignalWithInit interface {
	Signal

	Init(ctx workflow.Context) error
}

// ApplyInit checks if the signal implements SignalWithInit and calls Init if so.
func ApplyInit(sig Signal, ctx workflow.Context) error {
	if si, ok := sig.(SignalWithInit); ok {
		return si.Init(ctx)
	}
	return nil
}
