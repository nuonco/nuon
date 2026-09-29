package signal

import "go.temporal.io/sdk/workflow"

type SignalWithUpdateHandlers interface {
	Signal

	RegisterUpdateHandlers(ctx workflow.Context) error
}

func RegisterUpdateHandlers(sig Signal, ctx workflow.Context) error {
	if su, ok := sig.(SignalWithUpdateHandlers); ok {
		return su.RegisterUpdateHandlers(ctx)
	}

	return nil
}
