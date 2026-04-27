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

// ForwardStepFinishedRequest is the input for forwarding a step-finished update
// to a step handler workflow.
type ForwardStepFinishedRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// StepFinishedResponse is the typed response from the step-finished update
// handler. It contains the step's final status and directive so callers don't
// need to re-fetch the step from the database.
type StepFinishedResponse struct {
	StepID    string     `json:"step_id"`
	Status    app.Status `json:"status"`
	Directive string     `json:"directive"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @schedule-to-close-timeout 2h
// @heartbeat-timeout 60s
func (a *Activities) ForwardStepFinished(ctx context.Context, req ForwardStepFinishedRequest) (*StepFinishedResponse, error) {
	var qs app.QueueSignal
	res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.StepID,
			OwnerType: (&app.WorkflowStep{}).TableName(),
			Type:      "execute-workflow-step",
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find step queue signal for step %s: %w", req.StepID, res.Error)
	}

	return heartbeat.WithHeartbeat(ctx, 30*time.Second, func(ctx context.Context) (*StepFinishedResponse, error) {
		rawResp, err := handler.UpdateWithStart(ctx, a.tClient, &qs, handler.UpdateWithStartOptions{
			UpdateName:   "step-finished",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to send step-finished update to step %s: %w", req.StepID, err)
		}

		var resp StepFinishedResponse
		if err := rawResp.Get(ctx, &resp); err != nil {
			return nil, fmt.Errorf("step-finished update failed for step %s: %w", req.StepID, err)
		}

		return &resp, nil
	})
}
