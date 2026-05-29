package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

// sendCompletionCallbacks sends Temporal signals to all registered parent
// workflows with the handler's terminal status.
//
// It reloads the QueueSignal from the DB before sending so that callbacks
// added after initializeState (e.g. by EnsureSignal) are picked up.
func (h *handler) sendCompletionCallbacks(ctx workflow.Context) {
	l, _ := log.WorkflowLogger(ctx)

	// Reload from DB to pick up callbacks added after init (e.g. by EnsureSignal).
	qs, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, h.queueSignalID)
	if err == nil {
		h.callbacks = qs.Callbacks
		// Merge legacy single Callback if set.
		if qs.Callback.IsSet() {
			found := false
			for _, cb := range h.callbacks {
				if cb.WorkflowID == qs.Callback.WorkflowID && cb.SignalName == qs.Callback.SignalName {
					found = true
					break
				}
			}
			if !found {
				h.callbacks = append(h.callbacks, qs.Callback)
			}
		}
	}

	if !h.callbacks.IsSet() {
		return
	}

	result := callback.Result{
		Status:            string(h.finishedStatus),
		StatusDescription: h.finishedErr,
	}
	for _, cb := range h.callbacks {
		callback.Send(ctx, l, cb, result)
	}
}

// hasCallbacks returns true if at least one completion callback is configured.
func (h *handler) hasCallbacks() bool {
	return h.callbacks.IsSet()
}
