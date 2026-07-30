package worker

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/types/workflows/phonehomesecretbackfill"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	backfillPhoneHomeActivityTimeout = 5 * time.Minute
	// backfillPhoneHomeBatchesPerRun caps batches per workflow run before
	// continue-as-new, keeping workflow history bounded on large fleets.
	backfillPhoneHomeBatchesPerRun = 20
)

// BackfillPhoneHomeSecrets provisions phone-home secrets across the existing fleet,
// draining it in keyset-paginated batches and reconciling each install with the same
// EnsureInstallPhoneHomeSecret activity the provisioning path uses. It
// continue-as-news every backfillPhoneHomeBatchesPerRun batches so history stays
// bounded, and exposes a progress query. Idempotent and safe to re-run.
//
// This lives in the installs namespace rather than general because the reconciler
// activity is registered on the installs worker; a general-namespace workflow cannot
// reach it.
//
// Note this is pre-provisioning, not enforcement. It creates the secret, mints the
// tokens and applies the cross-account grant, but an already-deployed phone-home
// Lambda carries no NUON_PHONE_HOME_* environment variables and will not send a
// token until its stack version is regenerated and the customer re-applies.
func (w *Workflows) BackfillPhoneHomeSecrets(ctx workflow.Context, req phonehomesecretbackfill.Request) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	progress := phonehomesecretbackfill.Progress{
		InstallsProcessed: req.InstallsProcessed,
		SecretsEnsured:    req.SecretsEnsured,
		TokensMinted:      req.TokensMinted,
		Skipped:           req.Skipped,
		Errors:            req.Errors,
		CursorCreatedAt:   req.CursorCreatedAt,
		CursorID:          req.CursorID,
	}
	if err := workflow.SetQueryHandler(ctx, phonehomesecretbackfill.ProgressQueryType, func() (phonehomesecretbackfill.Progress, error) {
		return progress, nil
	}); err != nil {
		return errors.Wrap(err, "unable to register progress query handler")
	}

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = phonehomesecretbackfill.DefaultBatchSize
	}

	actx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: backfillPhoneHomeActivityTimeout,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	})

	for batches := 0; batches < backfillPhoneHomeBatchesPerRun; batches++ {
		page, err := activities.AwaitListPhoneHomeBackfillInstalls(actx, activities.ListPhoneHomeBackfillInstallsRequest{
			CursorCreatedAt: req.CursorCreatedAt,
			CursorID:        req.CursorID,
			Limit:           batchSize,
		})
		if err != nil {
			return errors.Wrap(err, "unable to list phone home backfill installs")
		}

		for _, installID := range page.InstallIDs {
			// One activity per install so Temporal retries each independently, and a
			// single bad install cannot fail the batch. A reconcile that still fails
			// after its retries is counted and stepped over — the cursor advances, so
			// the backfill drains rather than wedging on one install.
			resp, err := activities.AwaitEnsureInstallPhoneHomeSecretByInstallID(actx, installID)
			if err != nil {
				req.Errors++
				l.Warn("unable to ensure phone home secret during backfill",
					zap.String("install_id", installID), zap.Error(err))

				continue
			}

			switch {
			case resp.Skipped:
				req.Skipped++
			default:
				req.SecretsEnsured++
				req.TokensMinted += resp.TokensMinted
			}
		}

		req.InstallsProcessed += page.Examined
		if page.LastID != "" {
			req.CursorCreatedAt = page.LastCreatedAt
			req.CursorID = page.LastID
		}

		progress.InstallsProcessed = req.InstallsProcessed
		progress.SecretsEnsured = req.SecretsEnsured
		progress.TokensMinted = req.TokensMinted
		progress.Skipped = req.Skipped
		progress.Errors = req.Errors
		progress.CursorCreatedAt = req.CursorCreatedAt
		progress.CursorID = req.CursorID

		// A short batch means the fleet is drained.
		if page.Examined < batchSize {
			progress.Done = true
			l.Info("phone home secret backfill complete",
				zap.Int("installs_processed", req.InstallsProcessed),
				zap.Int("secrets_ensured", req.SecretsEnsured),
				zap.Int("tokens_minted", req.TokensMinted),
				zap.Int("skipped", req.Skipped),
				zap.Int("errors", req.Errors),
			)

			return nil
		}
	}

	l.Info("continuing phone home secret backfill",
		zap.Int("installs_processed", req.InstallsProcessed),
		zap.Time("cursor_created_at", req.CursorCreatedAt),
		zap.String("cursor_id", req.CursorID),
	)

	return workflow.NewContinueAsNewError(ctx, phonehomesecretbackfill.WorkflowName, req)
}
