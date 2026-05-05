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
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
		return terminalResponseFromMeta(q.Status.Status, q.Status.StatusHumanDescription, q.Status.Metadata)
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
				return terminalResponseFromMeta(fresh.Status.Status, fresh.Status.StatusHumanDescription, fresh.Status.Metadata)
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
				return terminalResponseFromMeta(fresh.Status.Status, fresh.Status.StatusHumanDescription, fresh.Status.Metadata)
			}
			return nil, errors.Wrap(err, "unable get response")
		}

		// The handler returned a terminal status directly - use it.
		if resp.Status == app.StatusError {
			return nil, signalFailedError(resp.StatusDescription, resp.Metadata)
		}

		return &resp, nil
	})
}

// terminalResponseFromMeta converts a terminal DB status into the appropriate return value.
//
// On error, the returned temporal NonRetryableApplicationError carries a
// stderr.StepErrorPayload as ApplicationError details when the original
// failure was an stderr.ErrUser (the metadata map carries the
// MetadataKeyErrorCode / MetadataKeyErrorFields / MetadataKeyStepDirective
// fields). Conductor code can recover the typed error via errors.As + the
// ApplicationError.Details API.
func terminalResponseFromMeta(status app.Status, description string, meta map[string]any) (*handler.FinishedResponse, error) {
	if status == app.StatusError {
		msg := description
		if msg == "" {
			msg = fmt.Sprintf("signal execution failed with status: %s", status)
		}
		return nil, signalFailedError(msg, meta)
	}
	return &handler.FinishedResponse{Status: status, StatusDescription: description, Metadata: meta}, nil
}

// signalFailedError builds the canonical SIGNAL_FAILED non-retryable
// ApplicationError. When meta carries StepErrorPayload-shaped fields, they
// are attached as details so the workflow side can recover code/fields/
// directive without re-reading the QueueSignal.
func signalFailedError(msg string, meta map[string]any) error {
	payload := stderr.PayloadFromMeta(meta)
	if payload.IsZero() {
		return temporal.NewNonRetryableApplicationError(msg, "SIGNAL_FAILED", nil)
	}
	return temporal.NewNonRetryableApplicationError(msg, "SIGNAL_FAILED", nil, payload)
}

func (c *Client) getQueueSignal(ctx context.Context, id string) (*app.QueueSignal, error) {
	var q app.QueueSignal
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get queue signal")
	}

	return &q, nil
}
