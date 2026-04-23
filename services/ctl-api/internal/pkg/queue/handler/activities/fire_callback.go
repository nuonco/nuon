package activities

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"
)

// FireQueueCallbackRequest is the input for the FireQueueCallback activity.
// Populated by the queue handler from the callback fields stored on the
// QueueSignal at enqueue time.
type FireQueueCallbackRequest struct {
	QueueSignalID string
	WorkflowID    string
	Namespace     string
	UpdateName    string
	Payload       CallbackPayload
}

// CallbackPayload is delivered to the caller's update handler when a queued
// signal finishes processing.
type CallbackPayload struct {
	QueueSignalID string `json:"queue_signal_id"`
	Success       bool   `json:"success"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// FireQueueCallback sends a workflow update to the caller that registered a
// callback on a queue signal. Waits only for the update to be accepted — we
// do not need the caller's handler return value.
//
// Failures here are non-fatal: the signal status is already persisted in the
// DB and the caller can fall back to AwaitSignal. The activity returns the
// error so the handler can log it.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @schedule-to-close-timeout 5m
func (a *Activities) FireQueueCallback(ctx context.Context, req FireQueueCallbackRequest) error {
	if req.WorkflowID == "" || req.UpdateName == "" {
		return nil
	}

	_, err := a.tclient.UpdateWorkflowInNamespace(ctx, req.Namespace, tclient.UpdateWorkflowOptions{
		UpdateID:     req.QueueSignalID + "-callback",
		WorkflowID:   req.WorkflowID,
		UpdateName:   req.UpdateName,
		WaitForStage: tclient.WorkflowUpdateStageAccepted,
		Args:         []any{req.Payload},
	})
	if err != nil {
		return errors.Wrap(err, "unable to send callback update")
	}
	return nil
}
