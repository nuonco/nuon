package client

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/pkg/temporal/heartbeat"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FetchStepsRequest struct {
	QueueSignalID string `json:"queue_signal_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
// @heartbeat-timeout 10s
func (c *Client) FetchSteps(ctx context.Context, req FetchStepsRequest) ([]*app.WorkflowStep, error) {
	q, err := c.getQueueSignal(ctx, req.QueueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	return heartbeat.WithHeartbeat(ctx, 3*time.Second, func(ctx context.Context) ([]*app.WorkflowStep, error) {
		rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   "FetchSteps",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to send FetchSteps update to signal %s: %w", req.QueueSignalID, err)
		}

		var steps []*app.WorkflowStep
		if err := rawResp.Get(ctx, &steps); err != nil {
			return nil, fmt.Errorf("FetchSteps update failed for signal %s: %w", req.QueueSignalID, err)
		}

		return steps, nil
	})
}
