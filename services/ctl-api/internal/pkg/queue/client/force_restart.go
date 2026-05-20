package client

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// ForceRestart terminates the running queue workflow and starts a fresh one.
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) ForceRestart(ctx context.Context, queueID string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	if q.OrgID != nil {
		ctx = cctx.SetOrgIDContext(ctx, *q.OrgID)
	}

	wkflowReq := queue.QueueWorkflowRequest{
		QueueID: q.ID,
		Version: c.cfg.Version,
	}
	opts := tclient.StartWorkflowOptions{
		ID:                    q.Workflow.ID,
		TaskQueue:             workflows.APITaskQueue,
		Memo:                  queueMemo(q),
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	wkflowRun, err := c.tClient.ExecuteWorkflowInNamespace(ctx,
		q.Workflow.Namespace,
		opts,
		"Queue",
		wkflowReq,
	)
	if err != nil {
		return errors.Wrap(err, "unable to force restart queue workflow")
	}

	c.l.Debug("queue force restarted",
		zap.String("namespace", q.Workflow.Namespace),
		zap.String("id", q.Workflow.ID),
		zap.String("run-id", wkflowRun.GetRunID()),
	)

	return nil
}
