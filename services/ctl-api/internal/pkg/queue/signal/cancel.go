package signal

import "go.temporal.io/sdk/workflow"

// SignalWithCancel is an optional interface that signals can implement to
// provide custom cancellation cleanup logic.
type SignalWithCancel interface {
	Signal

	Cancel(ctx workflow.Context) error
}
