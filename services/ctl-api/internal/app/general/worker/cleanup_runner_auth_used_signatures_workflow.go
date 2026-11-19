package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	cleanupRunnerAuthUsedSignatureBatchSize              = 5000
	cleanupRunnerAuthUsedSignatureMaxBatchesPerExecution = 500
	cleanupRunnerAuthUsedSignatureMaxRowsPerRun          = 10000000
)

type CleanupRunnerAuthUsedSignaturesRequest struct {
	TotalDeleted int64 `json:"total_deleted"`
}

func (w *Workflows) CleanupRunnerAuthUsedSignatures(ctx workflow.Context, req CleanupRunnerAuthUsedSignaturesRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("general workflow execution", zap.String("type", "cleanup-runner-auth-used-signatures-cron"))

	total := req.TotalDeleted
	for i := 0; i < cleanupRunnerAuthUsedSignatureMaxBatchesPerExecution; i++ {
		resp, err := activities.AwaitDeleteExpiredRunnerAuthUsedSignatures(ctx, activities.DeleteExpiredRunnerAuthUsedSignaturesRequest{
			BatchSize: cleanupRunnerAuthUsedSignatureBatchSize,
		})
		if err != nil {
			return errors.Wrap(err, "unable to delete expired runner auth used signatures")
		}

		total += resp.Deleted
		if resp.Deleted < cleanupRunnerAuthUsedSignatureBatchSize {
			l.Info("cleaned up expired runner auth used signatures", zap.Int64("deleted", total))
			return nil
		}
		if total >= cleanupRunnerAuthUsedSignatureMaxRowsPerRun {
			l.Warn("hit max rows cleaning up expired runner auth used signatures; backlog remains for next run", zap.Int64("deleted", total))
			return nil
		}
	}

	l.Info("continuing runner auth used signature cleanup in new execution", zap.Int64("deleted_so_far", total))
	return workflow.NewContinueAsNewError(ctx, w.CleanupRunnerAuthUsedSignatures, CleanupRunnerAuthUsedSignaturesRequest{TotalDeleted: total})
}
