package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	cleanupQueueSignalBatchSize = 5000
	// cap batches per execution so workflow history stays bounded; we Continue-As-New to keep draining.
	cleanupQueueSignalMaxBatchesPerExecution = 500
	// cap rows per cron trigger (across Continue-As-New executions); a larger backlog drains over subsequent daily runs.
	cleanupQueueSignalMaxRowsPerRun = 10000000
)

type CleanupQueueSignalsRequest struct {
	TotalDeleted int64 `json:"total_deleted"`
}

func (w *Workflows) CleanupQueueSignals(ctx workflow.Context, req CleanupQueueSignalsRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("general workflow execution", zap.String("type", "cleanup-queue-signals-cron"))

	total := req.TotalDeleted
	for i := 0; i < cleanupQueueSignalMaxBatchesPerExecution; i++ {
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
		if total >= cleanupQueueSignalMaxRowsPerRun {
			l.Warn("hit max rows cleaning up old queue signals; backlog remains for next run", zap.Int64("deleted", total))
			return nil
		}
	}

	l.Info("continuing queue signal cleanup in new execution", zap.Int64("deleted_so_far", total))
	return workflow.NewContinueAsNewError(ctx, w.CleanupQueueSignals, CleanupQueueSignalsRequest{TotalDeleted: total})
}
