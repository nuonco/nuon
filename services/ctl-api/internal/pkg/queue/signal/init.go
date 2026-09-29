package signal

import "go.temporal.io/sdk/workflow"

type SignalWithInit interface {
	Signal

	Init(ctx workflow.Context) error
}

func ApplyInit(sig Signal, ctx workflow.Context) error {
	if si, ok := sig.(SignalWithInit); ok {
		return si.Init(ctx)
	}
	return nil
}
