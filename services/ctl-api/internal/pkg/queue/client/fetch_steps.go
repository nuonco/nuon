package client

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/temporal/heartbeat"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

type FetchStepsRequest struct {
	QueueSignalID string `json:"queue_signal_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
// @heartbeat-timeout 60s
func (c *Client) FetchSteps(ctx context.Context, req FetchStepsRequest) (*app.GenerateStepsResult, error) {
	q, err := c.getQueueSignal(ctx, req.QueueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	// Recover run ID from: heartbeat (fastest) → DB (persisted by dispatcher) → Ready update (fallback).
	var runID string
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &runID); err != nil {
			return nil, errors.Wrap(err, "unable to get heartbeat details")
		}
	}

	if runID == "" && q.Workflow.RunID != "" {
		runID = q.Workflow.RunID
	}

	if runID == "" {
		resp, err := c.awaitHandlerReady(ctx, q)
		if err != nil {
			return nil, errors.Wrap(err, "handler not ready")
		}
		runID = resp.RunID
	}
	activity.RecordHeartbeat(ctx, runID)

	return heartbeat.WithHeartbeat(ctx, 30*time.Second, func(ctx context.Context) (*app.GenerateStepsResult, error) {
		rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
			UpdateID:     req.QueueSignalID + "-fetch-steps",
			WorkflowID:   q.Workflow.ID,
			RunID:        runID,
			UpdateName:   "FetchSteps",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
		if err != nil {
			// The update targets the original handler run that has the generated
			// steps in memory. If that run terminated, the steps are lost and
			// the caller must re-enqueue the generate-steps signal.
			return nil, temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("FetchSteps update failed for signal %s (run %s): %s", req.QueueSignalID, runID, err),
				"FETCH_STEPS_FAILED",
				err,
			)
		}

		var result app.GenerateStepsResult
		if err := rawResp.Get(ctx, &result); err != nil {
			return nil, temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("FetchSteps result failed for signal %s: %s", req.QueueSignalID, err),
				"FETCH_STEPS_FAILED",
				err,
			)
		}

		return &result, nil
	})
}

func (c *Client) awaitHandlerReady(ctx context.Context, q *app.QueueSignal) (*handler.ReadyResponse, error) {
	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		UpdateID:     q.ID + "-handler-ready",
		WorkflowID:   q.Workflow.ID,
		UpdateName:   handler.ReadyHandlerName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return nil, fmt.Errorf("ready update failed for signal %s: %w", q.ID, err)
	}
	var resp handler.ReadyResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("ready response failed for signal %s: %w", q.ID, err)
	}
	return &resp, nil
}
