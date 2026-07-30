package client

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// migrateQueueNamespace moves a queue's workflow from its current Temporal
// namespace to newNamespace. The old-namespace workflow is terminated
// (since its cross-namespace, TERMINATE_IF_RUNNING on the restart is not enough), the
// DB row is repointed, and the workflow is restarted in the new namespace.
//
//	Emitters are migrated separately by the emitter client (queue/client cannot import the emitter package).
func (c *Client) migrateQueueNamespace(ctx context.Context, q *app.Queue, newNamespace, newTaskQueue string) error {
	oldNamespace := q.Workflow.Namespace

	if c.tClient != nil {
		var nsClient tclient.Client
		var err error
		if nsClient, err = c.tClient.GetNamespaceClient(oldNamespace); err != nil {
			c.l.Warn("unable to get namespace client for migration terminate",
				zap.String("namespace", oldNamespace), zap.Error(err))
		} else if err := nsClient.TerminateWorkflow(
			ctx,
			q.Workflow.ID,
			"",
			"queue namespace migration",
		); err != nil {
			c.l.Warn("unable to terminate old queue workflow during namespace migration",
				zap.String("namespace", oldNamespace),
				zap.String("workflow-id", q.Workflow.ID),
				zap.Error(err),
			)
		}
	}

	q.Workflow.Namespace = newNamespace
	q.Workflow.TaskQueue = newTaskQueue
	if res := c.db.WithContext(ctx).Model(q).Update("workflow", q.Workflow); res.Error != nil {
		return errors.Wrap(res.Error, "unable to repoint queue row")
	}

	if c.tClient == nil {
		return nil
	}

	opts := tclient.StartWorkflowOptions{
		ID:                       q.Workflow.ID,
		TaskQueue:                newTaskQueue,
		Memo:                     queueMemo(q),
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
		RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 0},
	}
	if _, err := c.tClient.ExecuteWorkflowInNamespace(ctx, newNamespace, opts,
		"Queue", queue.QueueWorkflowRequest{QueueID: q.ID, Version: c.cfg.Version}); err != nil {
		return errors.Wrap(err, "unable to restart queue workflow in new namespace")
	}

	c.l.Info("migrated queue to new namespace",
		zap.String("queue-id", q.ID),
		zap.String("from", oldNamespace),
		zap.String("to", newNamespace),
	)
	return nil
}
