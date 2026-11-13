package worker

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	cleanupDeadTokensBatchSize  = 1000
	cleanupDeadTokensRunBudget  = 12 * time.Minute
	cleanupDeadTokensBatchPause = time.Second
)

func (w *Workflows) CleanupDeadTokens(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	var afterID string
	var selected, deleted int
	deadline := workflow.Now(ctx).Add(cleanupDeadTokensRunBudget)
	for workflow.Now(ctx).Before(deadline) {
		resp, err := activities.AwaitDeleteDeadTokens(ctx, activities.DeleteDeadTokensRequest{
			AfterID:   afterID,
			BatchSize: cleanupDeadTokensBatchSize,
		})
		if err != nil {
			return errors.Wrap(err, "unable to delete dead tokens")
		}
		if !resp.Enabled {
			return nil
		}
		selected += resp.Selected
		deleted += resp.Deleted
		if resp.Selected < cleanupDeadTokensBatchSize {
			break
		}
		afterID = resp.LastID
		if err := workflow.Sleep(ctx, cleanupDeadTokensBatchPause); err != nil {
			return err
		}
	}

	l.Info("cleaned up dead tokens",
		zap.Int("selected", selected),
		zap.Int("deleted", deleted))
	return nil
}
