package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/pkg/temporal/heartbeat"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @schedule-to-close-timeout 2h
// @heartbeat-timeout 10s
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowValidate(ctx context.Context, workflowID string, updateID string, queueID string) (*handler.ValidateResponse, error) {
	return heartbeat.WithHeartbeat(ctx, 3*time.Second, func(ctx context.Context) (*handler.ValidateResponse, error) {
		info := activity.GetInfo(ctx)

		rawResp, err := a.tclient.UpdateWithStartWorkflowInNamespace(ctx,
			info.WorkflowNamespace,
			tclient.UpdateWithStartWorkflowOptions{
				UpdateOptions: tclient.UpdateWorkflowOptions{
					UpdateID:     updateID + "-validate",
					WorkflowID:   workflowID,
					UpdateName:   handler.ValidateUpdateName,
					WaitForStage: tclient.WorkflowUpdateStageCompleted,
				},
				StartWorkflowOperation: a.handlerStartOperation(workflowID, queueID, updateID),
			})
		if err != nil {
			return nil, errors.Wrap(err, "unable to call query handler")
		}

		var resp handler.ValidateResponse
		if err := rawResp.Get(ctx, &resp); err != nil {
			return nil, errors.Wrap(err, "unable get response")
		}

		return &resp, nil
	})
}
