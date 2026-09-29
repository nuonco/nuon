package client

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

func (c *Client) QuerySignalStatus(ctx context.Context, queueSignalID string) (*handler.StatusResponse, error) {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	resp, err := c.tClient.QueryWorkflowInNamespace(ctx, q.Workflow.Namespace, q.Workflow.ID, "", handler.StatusQueryName, &handler.StatusRequest{})
	if err == nil {
		var status handler.StatusResponse
		if err := resp.Get(&status); err != nil {
			return nil, errors.Wrap(err, "unable to decode status response")
		}
		return &status, nil
	}

	return statusResponseFromDBStatus(q.Status), nil
}

func isTerminalStatus(s app.Status) bool {
	switch s {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled:
		return true
	default:
		return false
	}
}

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
		resp.Finished = true
	}

	return resp
}
