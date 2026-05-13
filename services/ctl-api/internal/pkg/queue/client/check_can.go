package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// CheckCAN triggers an on-demand CAN check on a queue workflow.
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) CheckCAN(ctx context.Context, queueID string) (*queue.CheckCANResponse, error) {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	if q.OrgID != nil {
		ctx = cctx.SetOrgIDContext(ctx, *q.OrgID)
	}

	rawResp, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.CheckCANUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				queue.CheckCANRequest{},
			},
		},
		StartWorkflowOperation: c.queueStartOperation(q),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call check-can handler")
	}

	var resp queue.CheckCANResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get check-can response")
	}

	return &resp, nil
}
