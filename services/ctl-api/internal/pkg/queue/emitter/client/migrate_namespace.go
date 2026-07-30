package client

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/taskqueue"
)

// MigrateQueueEmitters moves every emitter of a queue to newNamespace. Each
// old-namespace emitter workflow is terminated (cross-namespace, so
// TERMINATE_IF_RUNNING on the restart is not enough), its DB row repointed, and
// the workflow restarted in the new namespace. Emitters already in newNamespace
// are skipped, so this is a no-op when nothing changed.
func (c *Client) MigrateQueueEmitters(ctx context.Context, queueID string, newNamespace string) error {
	var emitters []app.QueueEmitter
	if res := c.db.WithContext(ctx).Where(app.QueueEmitter{QueueID: queueID}).Find(&emitters); res.Error != nil {
		return errors.Wrap(res.Error, "unable to load emitters for migration")
	}

	newTaskQueue := taskqueue.For(newNamespace, "")

	for i := range emitters {
		em := &emitters[i]
		if em.Workflow.Namespace == newNamespace {
			continue
		}
		oldNamespace := em.Workflow.Namespace

		if c.tClient != nil {
			if nsClient, err := c.tClient.GetNamespaceClient(oldNamespace); err != nil {
				c.l.Warn("unable to get namespace client for emitter migration terminate",
					zap.String("namespace", oldNamespace), zap.Error(err))
			} else if err := nsClient.TerminateWorkflow(ctx, em.Workflow.ID, "", "queue namespace migration"); err != nil {
				c.l.Warn("unable to terminate old emitter workflow during namespace migration",
					zap.String("emitter-id", em.ID), zap.Error(err))
			}
		}

		em.Workflow.Namespace = newNamespace
		em.Workflow.TaskQueue = newTaskQueue
		if res := c.db.WithContext(ctx).Model(em).Update("workflow", em.Workflow); res.Error != nil {
			return errors.Wrapf(res.Error, "unable to repoint emitter %s", em.ID)
		}

		if c.tClient == nil {
			continue
		}

		opts := tclient.StartWorkflowOptions{
			ID:                       em.Workflow.ID,
			TaskQueue:                newTaskQueue,
			Memo:                     emitterMemo(em),
			WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
			WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
			RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 0},
		}
		if em.Mode == app.QueueEmitterModeCron {
			opts.CronSchedule = em.CronSchedule
		}
		if _, err := c.tClient.ExecuteWorkflowInNamespace(ctx, newNamespace, opts,
			"Emitter", emitter.EmitterWorkflowRequest{QueueID: queueID, EmitterID: em.ID, Version: c.cfg.Version}); err != nil {
			return errors.Wrapf(err, "unable to restart emitter %s in new namespace", em.ID)
		}

		c.l.Info("migrated emitter to new namespace",
			zap.String("emitter-id", em.ID),
			zap.String("from", oldNamespace),
			zap.String("to", newNamespace),
		)
	}

	return nil
}
