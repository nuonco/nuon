package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	sweepLogStreamAccountsBatchSize  = 1000
	sweepLogStreamAccountsMaxBatches = 20
)

func (w *Workflows) SweepLogStreamAccounts(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	var afterID string
	var selected, deleted, skipped int
	for i := 0; i < sweepLogStreamAccountsMaxBatches; i++ {
		resp, err := activities.AwaitSweepLogStreamAccounts(ctx, activities.SweepLogStreamAccountsRequest{
			AfterID:   afterID,
			BatchSize: sweepLogStreamAccountsBatchSize,
		})
		if err != nil {
			return errors.Wrap(err, "unable to sweep log stream accounts")
		}
		if !resp.Enabled {
			return nil
		}
		selected += resp.Selected
		deleted += resp.Deleted
		skipped += resp.Skipped
		if resp.Selected < sweepLogStreamAccountsBatchSize {
			break
		}
		afterID = resp.LastID
	}

	l.Info("swept log stream accounts",
		zap.Int("selected", selected),
		zap.Int("deleted", deleted),
		zap.Int("skipped", skipped))
	return nil
}
