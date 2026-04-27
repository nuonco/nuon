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
		if q.Status.Status == app.StatusError {
			return nil, temporal.NewNonRetryableApplicationError(
				signalErrorMessage(q),
				"SIGNAL_FAILED", nil)
		}
		return &handler.FinishedResponse{}, nil
	}

	return heartbeat.WithHeartbeat(ctx, 30*time.Second, func(ctx context.Context) (*handler.FinishedResponse, error) {
		rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
			UpdateID:     queueSignalID + "-finished",
			WorkflowID:   q.Workflow.ID,
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
				if fresh.Status.Status == app.StatusError {
					return nil, temporal.NewNonRetryableApplicationError(
						signalErrorMessage(fresh),
						"SIGNAL_FAILED", nil)
				}
				return &handler.FinishedResponse{}, nil
			}
			return nil, errors.Wrap(err, "unable to call finished handler")
		}

		var resp handler.FinishedResponse
		if err := rawResp.Get(ctx, &resp); err != nil {
			return nil, errors.Wrap(err, "unable get response")
		}

		// Re-check DB status after workflow update completes, since the finishedHandler
		// returns FinishedResponse{} regardless of whether signal execution succeeded or failed.
		fresh, err := c.getQueueSignal(ctx, queueSignalID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to verify signal status after completion")
		}
		if fresh.Status.Status == app.StatusError {
			return nil, temporal.NewNonRetryableApplicationError(
				signalErrorMessage(fresh),
				"SIGNAL_FAILED", nil)
		}

		return &resp, nil
	})
}

func signalErrorMessage(q *app.QueueSignal) string {
	if q.Status.StatusHumanDescription != "" {
		return q.Status.StatusHumanDescription
	}
	return fmt.Sprintf("signal execution failed with status: %s", q.Status.Status)
}

func (c *Client) getQueueSignal(ctx context.Context, id string) (*app.QueueSignal, error) {
	var q app.QueueSignal
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get queue signal")
	}

	return &q, nil
}
