package activities

import (
	"context"
	"fmt"
	"time"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/pkg/temporal/heartbeat"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// ForwardGroupFinishedRequest is the input for forwarding a group-finished
// update to a group handler workflow.
type ForwardGroupFinishedRequest struct {
	StepGroupID string `json:"step_group_id" validate:"required"`
}

// GroupFinishedResponse is the typed response from the group-finished update
// handler. It contains the group's final directive so callers don't need to
// re-fetch the workflow's ResultDirective from the database.
type GroupFinishedResponse struct {
	Directive string `json:"directive"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @schedule-to-close-timeout 2h
// @heartbeat-timeout 60s
func (a *Activities) ForwardGroupFinished(ctx context.Context, req ForwardGroupFinishedRequest) (*GroupFinishedResponse, error) {
	var qs app.QueueSignal
	res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.StepGroupID,
			OwnerType: (&app.WorkflowStepGroup{}).TableName(),
			Type:      "execute-workflow-step-group",
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find group queue signal for group %s: %w", req.StepGroupID, res.Error)
	}

	return heartbeat.WithHeartbeat(ctx, 30*time.Second, func(ctx context.Context) (*GroupFinishedResponse, error) {
		rawResp, err := handler.UpdateWithStart(ctx, a.tClient, &qs, handler.UpdateWithStartOptions{
			UpdateName:   "group-finished",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to send group-finished update to group %s: %w", req.StepGroupID, err)
		}

		var resp GroupFinishedResponse
		if err := rawResp.Get(ctx, &resp); err != nil {
			return nil, fmt.Errorf("group-finished update failed for group %s: %w", req.StepGroupID, err)
		}

		return &resp, nil
	})
}
