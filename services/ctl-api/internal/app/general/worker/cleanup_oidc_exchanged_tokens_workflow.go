package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	cleanupOIDCExchangedTokenBatchSize = 5000
	// cap batches per execution so workflow history stays bounded; we Continue-As-New to keep draining.
	cleanupOIDCExchangedTokenMaxBatchesPerExecution = 500
	// cap rows per cron trigger (across Continue-As-New executions); a larger backlog drains over subsequent daily runs.
	cleanupOIDCExchangedTokenMaxRowsPerRun = 10000000
)

type CleanupOIDCExchangedTokensRequest struct {
	TotalDeleted int64 `json:"total_deleted"`
}

func (w *Workflows) CleanupOIDCExchangedTokens(ctx workflow.Context, req CleanupOIDCExchangedTokensRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("general workflow execution", zap.String("type", "cleanup-oidc-exchanged-tokens-cron"))

	total := req.TotalDeleted
	for i := 0; i < cleanupOIDCExchangedTokenMaxBatchesPerExecution; i++ {
		resp, err := activities.AwaitDeleteOldOIDCExchangedTokens(ctx, activities.DeleteOldOIDCExchangedTokensRequest{
			BatchSize: cleanupOIDCExchangedTokenBatchSize,
		})
		if err != nil {
			return errors.Wrap(err, "unable to delete old oidc exchanged tokens")
		}

		total += resp.Deleted
		if resp.Deleted < cleanupOIDCExchangedTokenBatchSize {
			l.Info("cleaned up old oidc exchanged tokens", zap.Int64("deleted", total))
			return nil
		}
		if total >= cleanupOIDCExchangedTokenMaxRowsPerRun {
			l.Warn("hit max rows cleaning up old oidc exchanged tokens; backlog remains for next run", zap.Int64("deleted", total))
			return nil
		}
	}

	l.Info("continuing oidc exchanged token cleanup in new execution", zap.Int64("deleted_so_far", total))
	return workflow.NewContinueAsNewError(ctx, w.CleanupOIDCExchangedTokens, CleanupOIDCExchangedTokensRequest{TotalDeleted: total})
}
