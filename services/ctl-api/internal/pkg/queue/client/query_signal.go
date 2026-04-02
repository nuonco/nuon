package client

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// QuerySignalStatus queries the status of a signal handler. If the handler workflow
// is sleeping or completed, it falls back to the database status.
func (c *Client) QuerySignalStatus(ctx context.Context, queueSignalID string) (*handler.StatusResponse, error) {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	// Try to query the existing workflow first
	resp, err := c.tClient.QueryWorkflowInNamespace(ctx, q.Workflow.Namespace, q.Workflow.ID, "", handler.StatusQueryName, &handler.StatusRequest{})
	if err == nil {
		var status handler.StatusResponse
		if err := resp.Get(&status); err != nil {
			return nil, errors.Wrap(err, "unable to decode status response")
		}
		return &status, nil
	}

	// Workflow query failed (handler is sleeping/completed) — fall back to DB status.
	return statusResponseFromDBStatus(q.Status), nil
}

// isTerminalStatus returns true if the DB status indicates the signal has finished processing.
func isTerminalStatus(s app.Status) bool {
	switch s {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled:
		return true
	default:
		return false
	}
}

// statusResponseFromDBStatus maps a persisted CompositeStatus to a handler.StatusResponse.
func statusResponseFromDBStatus(status app.CompositeStatus) *handler.StatusResponse {
	resp := &handler.StatusResponse{}

	switch status.Status {
	case app.StatusSuccess:
		resp.Finished = true
	case app.StatusError:
		resp.Finished = true
	case app.StatusCancelled:
		resp.Finished = true
		resp.Canceled = true
	default:
		// For queued/in-progress/etc the handler was sleeping before completion
		// which shouldn't normally happen, but treat as finished since the
		// workflow is gone.
		resp.Finished = true
		resp.Sleeping = true
	}

	return resp
}
