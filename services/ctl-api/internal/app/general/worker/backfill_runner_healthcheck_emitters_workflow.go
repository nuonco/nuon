package worker

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/types/workflows/runnerhealthcheckbackfill"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	backfillEmitterActivityTimeout = 5 * time.Minute
	// backfillEmitterBatchesPerRun caps batches per workflow run before
	// continue-as-new, keeping workflow history bounded on large fleets.
	backfillEmitterBatchesPerRun = 20
)

// BackfillRunnerHealthcheckEmitters drains the runner fleet in keyset-paginated
// batches, ensuring each runner has a runner-healthcheck emitter (creating and
// starting the emitter workflow for the ones that don't). It continue-as-news
// every backfillEmitterBatchesPerRun batches so history stays bounded, and
// exposes a progress query. Idempotent and safe to re-run.
func (w *Workflows) BackfillRunnerHealthcheckEmitters(ctx workflow.Context, req runnerhealthcheckbackfill.Request) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	progress := runnerhealthcheckbackfill.Progress{
		RunnersProcessed: req.RunnersProcessed,
		EmittersCreated:  req.EmittersCreated,
		AlreadyPresent:   req.AlreadyPresent,
		Errors:           req.Errors,
		CursorCreatedAt:  req.CursorCreatedAt,
		CursorID:         req.CursorID,
	}
	if err := workflow.SetQueryHandler(ctx, runnerhealthcheckbackfill.ProgressQueryType, func() (runnerhealthcheckbackfill.Progress, error) {
		return progress, nil
	}); err != nil {
		return errors.Wrap(err, "unable to register progress query handler")
	}

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = runnerhealthcheckbackfill.DefaultBatchSize
	}

	actx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: backfillEmitterActivityTimeout,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	})

	for batches := 0; batches < backfillEmitterBatchesPerRun; batches++ {
		var resp activities.BackfillRunnerHealthcheckEmittersResponse
		if err := workflow.ExecuteActivity(actx,
			(*activities.Activities).BackfillRunnerHealthcheckEmitters,
			activities.BackfillRunnerHealthcheckEmittersRequest{
				CursorCreatedAt: req.CursorCreatedAt,
				CursorID:        req.CursorID,
				Limit:           batchSize,
			},
		).Get(actx, &resp); err != nil {
			return errors.Wrap(err, "unable to backfill runner healthcheck emitters batch")
		}

		req.RunnersProcessed += resp.Examined
		req.EmittersCreated += resp.EmittersCreated
		req.AlreadyPresent += resp.AlreadyPresent
		req.Errors += len(resp.Errors)
		if resp.LastID != "" {
			req.CursorCreatedAt = resp.LastCreatedAt
			req.CursorID = resp.LastID
		}

		progress.RunnersProcessed = req.RunnersProcessed
		progress.EmittersCreated = req.EmittersCreated
		progress.AlreadyPresent = req.AlreadyPresent
		progress.Errors = req.Errors
		progress.CursorCreatedAt = req.CursorCreatedAt
		progress.CursorID = req.CursorID

		// A short batch means the fleet is drained.
		if resp.Examined < batchSize {
			progress.Done = true
			l.Info("runner healthcheck emitter backfill complete",
				zap.Int("runners_processed", req.RunnersProcessed),
				zap.Int("emitters_created", req.EmittersCreated),
				zap.Int("already_present", req.AlreadyPresent),
				zap.Int("errors", req.Errors),
			)
			return nil
		}
	}

	l.Info("continuing runner healthcheck emitter backfill",
		zap.Int("runners_processed", req.RunnersProcessed),
		zap.Time("cursor_created_at", req.CursorCreatedAt),
		zap.String("cursor_id", req.CursorID),
	)
	return workflow.NewContinueAsNewError(ctx, runnerhealthcheckbackfill.WorkflowName, req)
}
