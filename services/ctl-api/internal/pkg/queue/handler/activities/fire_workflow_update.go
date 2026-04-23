package activities

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"
)

// FireWorkflowUpdateRequest is the input for the FireWorkflowUpdate activity.
// Used by the queue handler's callback mechanism and by non-queue callers
// (e.g. ProcessJob push-update pattern) to fan out a workflow update to a
// specific target workflow.
type FireWorkflowUpdateRequest struct {
	// WorkflowID is the target workflow to update. No-op if empty.
	WorkflowID string
	// Namespace of the target workflow.
	Namespace string
	// UpdateName must match a handler registered via SetUpdateHandler on the
	// target workflow. No-op if empty.
	UpdateName string
	// UpdateID is optional. When set, Temporal deduplicates retries of the
	// same update. Leave empty for fire-and-forget semantics where each
	// retry is a distinct notification.
	UpdateID string
	// Payload is passed as the single argument to the target's update
	// handler. Any type serializable by Temporal's data converter.
	Payload any
}

// CallbackPayload is delivered to a queue-signal callback handler when the
// queued signal finishes processing. This is the payload shape for the
// queue handler's use of FireWorkflowUpdate.
type CallbackPayload struct {
	QueueSignalID string `json:"queue_signal_id"`
	Success       bool   `json:"success"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// FireWorkflowUpdate sends a workflow update to a target workflow. Waits
// only for the update to be accepted — we do not need the target's handler
// return value.
//
// Failures are non-fatal for the calling workflow to handle — log and
// continue. The target workflow may have completed/terminated, or the
// update handler may not be registered yet.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @schedule-to-close-timeout 5m
func (a *Activities) FireWorkflowUpdate(ctx context.Context, req FireWorkflowUpdateRequest) error {
	if req.WorkflowID == "" || req.UpdateName == "" {
		return nil
	}

	_, err := a.tclient.UpdateWorkflowInNamespace(ctx, req.Namespace, tclient.UpdateWorkflowOptions{
		UpdateID:     req.UpdateID,
		WorkflowID:   req.WorkflowID,
		UpdateName:   req.UpdateName,
		WaitForStage: tclient.WorkflowUpdateStageAccepted,
		Args:         []any{req.Payload},
	})
	if err != nil {
		return errors.Wrap(err, "unable to send workflow update")
	}
	return nil
}
