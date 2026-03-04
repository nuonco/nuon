package activities

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowReady(ctx context.Context, workflowID string, updateID string) (*handler.ReadyResponse, error) {
	info := activity.GetInfo(ctx)

	rawResp, err := a.tclient.UpdateWorkflowInNamespace(ctx,
		info.WorkflowNamespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   workflowID,
			UpdateName:   handler.ReadyHandlerName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call query handler")
	}

	var resp handler.ReadyResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}
