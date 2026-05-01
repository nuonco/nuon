package activities

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowReady(ctx context.Context, workflowID string, updateID string, queueID string) (*handler.ReadyResponse, error) {
	info := activity.GetInfo(ctx)

	startOp := a.handlerStartOperation(workflowID, queueID, updateID)
	rawResp, err := a.tclient.UpdateWithStartWorkflowInNamespace(ctx,
		info.WorkflowNamespace,
		tclient.UpdateWithStartWorkflowOptions{
			UpdateOptions: tclient.UpdateWorkflowOptions{
				UpdateID:     updateID + "-ready",
				WorkflowID:   workflowID,
				UpdateName:   handler.ReadyHandlerName,
				WaitForStage: tclient.WorkflowUpdateStageCompleted,
			},
			StartWorkflowOperation: startOp,
		})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call query handler")
	}

	var resp handler.ReadyResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		wrapped := errors.Wrap(err, "unable get response")
		var appErr *temporal.ApplicationError
		if errors.As(err, &appErr) && appErr.Type() == "AcceptedUpdateCompletedWorkflow" {
			return nil, temporal.NewNonRetryableApplicationError(
				appErr.Message(),
				appErr.Type(),
				wrapped,
			)
		}
		return nil, wrapped
	}

	// Resolve the run-id so callers can pin subsequent updates to this exact run.
	run, err := startOp.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get handler workflow run")
	}
	resp.RunID = run.GetRunID()

	return &resp, nil
}
