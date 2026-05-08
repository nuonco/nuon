package client

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// AwaitQueueSignal is a workflow-level helper that fetches the QueueSignal from
// the database, derives the appropriate timeout from the signal, and awaits its
// completion. This should be used instead of calling AwaitAwaitSignal directly
// so that every await uses the signal's declared timeout rather than a hardcoded
// default.
func AwaitQueueSignal(ctx workflow.Context, queueSignalID string) (*handler.FinishedResponse, error) {
	qs, err := AwaitGetQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, err
	}

	timeout := signal.DefaultTimeout
	if qs.Signal.Signal != nil {
		timeout = signal.DeriveTimeout(qs.Signal.Signal)
	}

	return AwaitAwaitSignal(ctx, queueSignalID, signal.TimeoutActivityOpts(timeout))
}
