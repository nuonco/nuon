package handler

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// sendCompletionCallbacks sends Temporal signals to all registered parent
// workflows with the handler's terminal status.
//
// It reloads the QueueSignal from the DB before sending so that callbacks
// added after initializeState (e.g. by EnsureSignal) are picked up.
func (h *handler) sendCompletionCallbacks(ctx workflow.Context) {
	l, _ := log.WorkflowLogger(ctx)

	if workflowID := completionCallbacksWorkflowID(h.sig); workflowID != "" {
		hold, err := activities.LocalAwaitHoldCompletionCallbacksByWorkflowID(ctx, workflowID)
		if err != nil {
			l.Error("unable to reload workflow before sending completion callbacks",
				zap.String("workflow_id", workflowID),
				zap.Error(err))
		} else if hold {
			return
		}
	}

	// Reload from DB to pick up callbacks added after init (e.g. by EnsureSignal).
	qs, err := activities.LocalAwaitGetQueueSignalByQueueSignalID(ctx, h.queueSignalID)
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

func completionCallbacksWorkflowID(sig signal.Signal) string {
	residentFlow, ok := sig.(signal.CompletionCallbacksWorkflow)
	if !ok {
		return ""
	}
	return residentFlow.CompletionCallbacksWorkflowID()
}

// hasCallbacks returns true if at least one completion callback is configured.
func (h *handler) hasCallbacks() bool {
	return h.callbacks.IsSet()
}
