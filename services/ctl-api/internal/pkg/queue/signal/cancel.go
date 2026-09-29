package signal

import "go.temporal.io/sdk/workflow"

type SignalWithCancel interface {
	Signal

	Cancel(ctx workflow.Context) error
}
