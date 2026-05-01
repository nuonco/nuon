package client

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/temporal/heartbeat"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @schedule-to-close-timeout 2h
// @heartbeat-timeout 60s
func (c *Client) AwaitSignal(ctx context.Context, queueSignalID string) (*handler.FinishedResponse, error) {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	// If the DB status already indicates completion, return immediately.
	if isTerminalStatus(q.Status.Status) {
		return terminalResponse(q.Status.Status, q.Status.StatusHumanDescription)
	}

	return heartbeat.WithHeartbeat(ctx, 30*time.Second, func(ctx context.Context) (*handler.FinishedResponse, error) {
		rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
			UpdateID:     queueSignalID + "-finished",
			WorkflowID:   q.Workflow.ID,
			RunID:        q.Workflow.RunID, // empty for old rows = latest run
			UpdateName:   handler.FinishedHandlerName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				handler.FinishedRequest{},
			},
		})
		if err != nil {
			// Workflow may have been slept/terminated between our DB check and now.
			// Re-check DB status to confirm.
			fresh, dbErr := c.getQueueSignal(ctx, queueSignalID)
			if dbErr != nil {
				return nil, errors.Wrap(dbErr, "unable to get queue signal from db")
			}
			if isTerminalStatus(fresh.Status.Status) {
				return terminalResponse(fresh.Status.Status, fresh.Status.StatusHumanDescription)
			}
			return nil, errors.Wrap(err, "unable to call finished handler")
		}

		var resp handler.FinishedResponse
		if err := rawResp.Get(ctx, &resp); err != nil {
			// The update itself may have returned an error. Check DB as fallback.
			fresh, dbErr := c.getQueueSignal(ctx, queueSignalID)
			if dbErr != nil {
				return nil, errors.Wrap(dbErr, "unable to get queue signal from db")
			}
			if isTerminalStatus(fresh.Status.Status) {
				return terminalResponse(fresh.Status.Status, fresh.Status.StatusHumanDescription)
			}
			return nil, errors.Wrap(err, "unable get response")
		}

		// The handler returned a terminal status directly - use it.
		if resp.Status == app.StatusError {
			return nil, temporal.NewNonRetryableApplicationError(
				resp.StatusDescription,
				"SIGNAL_FAILED", nil)
		}

		return &resp, nil
	})
}

// terminalResponse converts a terminal DB status into the appropriate return value.
func terminalResponse(status app.Status, description string) (*handler.FinishedResponse, error) {
	if status == app.StatusError {
		msg := description
		if msg == "" {
			msg = fmt.Sprintf("signal execution failed with status: %s", status)
		}
		return nil, temporal.NewNonRetryableApplicationError(msg, "SIGNAL_FAILED", nil)
	}
	return &handler.FinishedResponse{Status: status, StatusDescription: description}, nil
}

func (c *Client) getQueueSignal(ctx context.Context, id string) (*app.QueueSignal, error) {
	var q app.QueueSignal
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get queue signal")
	}

	return &q, nil
}
