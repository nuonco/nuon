package activities

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowSleep(ctx context.Context, workflowID string, updateID string) (*handler.SleepResponse, error) {
	info := activity.GetInfo(ctx)

	rawResp, err := a.tclient.UpdateWorkflowInNamespace(ctx,
		info.WorkflowNamespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   workflowID,
			UpdateName:   handler.SleepUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call sleep handler")
	}

	var resp handler.SleepResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}
