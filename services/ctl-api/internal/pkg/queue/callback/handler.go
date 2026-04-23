package callback

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// Handler registers a single Temporal update handler on the current workflow.
// The queue system invokes this update for any registered callback event,
// delivering a CallbackPayload that includes the event type and optional error.
//
// A Handler is tied to a single update name, but can be registered for multiple
// callback events (e.g. both OnSuccess and OnError). The caller uses Await or
// AwaitExecute to block until any registered event fires.
type Handler struct {
	ref signaldb.UpdateHandler
	ch  workflow.Channel
}

// NewHandler registers a Temporal update handler on the current workflow.
// The updateName should be unique within the workflow (e.g. "deploy-callback").
func NewHandler(ctx workflow.Context, updateName string) (*Handler, error) {
	info := workflow.GetInfo(ctx)

	h := &Handler{
		ref: signaldb.UpdateHandler{
			Namespace:  info.Namespace,
			WorkflowID: info.WorkflowExecution.ID,
			UpdateName: updateName,
		},
		ch: workflow.NewChannel(ctx),
	}

	if err := workflow.SetUpdateHandler(ctx, updateName, h.handleUpdate); err != nil {
		return nil, fmt.Errorf("unable to register callback handler %q: %w", updateName, err)
	}

	return h, nil
}

// Ref returns the UpdateHandler reference for a specific event. Use this when
// building CallbackRequest entries to pass to EnqueueSignal.
func (h *Handler) Ref() signaldb.UpdateHandler {
	return h.ref
}

// Callbacks returns CallbackRequest entries for the given events, all pointing
// at this handler's single update endpoint.
func (h *Handler) Callbacks(events ...Event) []CallbackRequest {
	out := make([]CallbackRequest, len(events))
	for i, e := range events {
		out[i] = CallbackRequest{
			Event:         e,
			UpdateHandler: h.ref,
		}
	}
	return out
}

// Await blocks until any registered callback fires and returns the payload.
func (h *Handler) Await(ctx workflow.Context) (CallbackPayload, error) {
	var payload CallbackPayload
	h.ch.Receive(ctx, &payload)
	if ctx.Err() != nil {
		return payload, ctx.Err()
	}
	return payload, nil
}

// AwaitExecute is a convenience method that blocks until an OnSuccess or OnError
// callback fires. Returns nil on success, or an error containing the error message
// from the queue signal execution on failure.
func (h *Handler) AwaitExecute(ctx workflow.Context) error {
	payload, err := h.Await(ctx)
	if err != nil {
		return err
	}

	if payload.Event == OnError {
		if payload.ErrMessage != "" {
			return fmt.Errorf("queue signal failed: %s", payload.ErrMessage)
		}
		return fmt.Errorf("queue signal failed")
	}

	return nil
}

// handleUpdate is the Temporal update handler that receives the callback payload
// and sends it to the waiting channel.
func (h *Handler) handleUpdate(ctx workflow.Context, payload CallbackPayload) error {
	h.ch.Send(ctx, payload)
	return nil
}
