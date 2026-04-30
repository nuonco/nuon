package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/temporal/heartbeat"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @schedule-to-close-timeout 2h
// @heartbeat-timeout 60s
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowValidate(ctx context.Context, workflowID string, updateID string, queueID string, runID string) (*handler.ValidateResponse, error) {
	return heartbeat.WithHeartbeat(ctx, 30*time.Second, func(ctx context.Context) (*handler.ValidateResponse, error) {
		info := activity.GetInfo(ctx)

		rawResp, err := a.tclient.UpdateWorkflowInNamespace(ctx,
			info.WorkflowNamespace,
			tclient.UpdateWorkflowOptions{
				UpdateID:     updateID + "-validate",
				WorkflowID:   workflowID,
				RunID:        runID,
				UpdateName:   handler.ValidateUpdateName,
				WaitForStage: tclient.WorkflowUpdateStageCompleted,
			})
		if err != nil {
			return nil, errors.Wrap(err, "unable to call update handler")
		}

		var resp handler.ValidateResponse
		if err := rawResp.Get(ctx, &resp); err != nil {
			wrapped := errors.Wrap(err, "unable get response")
			var appErr *temporal.ApplicationError
			if !errors.As(err, &appErr) {
				return nil, wrapped
			}
			if !appErr.NonRetryable() {
				return nil, wrapped
			}
			// Preserve non-retryable errors from the handler (signal failures,
			// AcceptedUpdateCompletedWorkflow, etc.) so Temporal stops retrying.
			return nil, temporal.NewNonRetryableApplicationError(
				wrapped.Error(),
				appErr.Type(),
				wrapped,
			)
		}

		return &resp, nil
	})
}
