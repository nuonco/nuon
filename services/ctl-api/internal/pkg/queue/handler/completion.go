package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// sendCompletionCallbacks sends a Temporal signal to the parent workflow
// (if configured) with the handler's terminal status.
func (h *handler) sendCompletionCallbacks(ctx workflow.Context) {
	if !h.callback.IsSet() {
		return
	}

	l, _ := log.WorkflowLogger(ctx)

	callback.Send(ctx, l, h.callback, callback.Result{
		Status:            string(h.finishedStatus),
		StatusDescription: h.finishedErr,
	})
}

// hasCallbacks returns true if a completion callback is configured.
func (h *handler) hasCallbacks() bool {
	return h.callback.IsSet()
}
