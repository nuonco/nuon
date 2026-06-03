package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	cleanupQueueSignalBatchSize = 50000
	// cap batches per run so workflow history stays bounded; a backlog is drained over subsequent daily runs.
	cleanupQueueSignalMaxBatches = 500
)

func (w *Workflows) CleanupQueueSignals(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("general workflow execution", zap.String("type", "cleanup-queue-signals-cron"))

	var total int64
	for i := 0; i < cleanupQueueSignalMaxBatches; i++ {
		resp, err := activities.AwaitDeleteOldQueueSignals(ctx, activities.DeleteOldQueueSignalsRequest{
			BatchSize: cleanupQueueSignalBatchSize,
		})
		if err != nil {
			return errors.Wrap(err, "unable to delete old queue signals")
		}

		total += resp.Deleted
		if resp.Deleted < cleanupQueueSignalBatchSize {
			l.Info("cleaned up old queue signals", zap.Int64("deleted", total))
			return nil
		}
	}

	l.Warn("hit max batches cleaning up old queue signals; backlog remains for next run", zap.Int64("deleted", total))
	return nil
}
