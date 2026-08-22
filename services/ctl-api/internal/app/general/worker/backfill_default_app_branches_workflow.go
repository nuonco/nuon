package worker

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/types/workflows/defaultappbranches"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	// caps apps per run so workflow history stays bounded across a fleet-wide backfill.
	defaultAppBranchesPerRun = 200

	// a permanently broken app must not hold the rest of the fleet for the
	// activity's 24h schedule-to-close default.
	defaultAppBranchAttempts = 3
	defaultAppBranchTimeout  = 5 * time.Minute

	// keeps the progress query readable when a whole class of apps fails.
	defaultAppBranchMaxFailedIDs = 50
)

// BackfillDefaultAppBranches gives every app the `default` branch and
// all-installs group that `nuon apps sync` otherwise creates lazily on its first
// run under default-app-branches, so flipping the flag does not make the next
// sync of each app a migration.
//
// One app failing is counted rather than fatal: a fleet-wide backfill that stops
// at the first bad app leaves the rest unmigrated with nothing to show for it.
func (w *Workflows) BackfillDefaultAppBranches(ctx workflow.Context, req defaultappbranches.Request) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	progress := defaultappbranches.Progress{
		DryRun:            req.DryRun,
		AppsTotal:         req.AppsTotal,
		AppsDone:          req.Created + req.Existing + req.Claimed + req.Failed,
		Created:           req.Created,
		Existing:          req.Existing,
		Claimed:           req.Claimed,
		Failed:            req.Failed,
		FailedAppIDs:      req.FailedAppIDs,
		InstallsConnected: req.InstallsConnected,
	}
	if err := workflow.SetQueryHandler(ctx, defaultappbranches.ProgressQueryType, func() (defaultappbranches.Progress, error) {
		return progress, nil
	}); err != nil {
		return errors.Wrap(err, "unable to register default app branch backfill progress query handler")
	}

	if !req.Initialized {
		l.Info("general workflow execution", zap.String("type", "default-app-branch-backfill"))

		resp, err := activities.AwaitListAppsNeedingDefaultBranch(ctx, activities.ListAppsNeedingDefaultBranchRequest{
			OrgIDs: req.OrgIDs,
		})
		if err != nil {
			return errors.Wrap(err, "unable to list apps needing a default branch")
		}

		req.Pending = resp.AppIDs
		req.AppsTotal = len(resp.AppIDs)
		req.Initialized = true
		progress.AppsTotal = req.AppsTotal
	}

	if req.DryRun {
		progress.Done = true
		l.Info("default app branch backfill dry run", zap.Int("apps_total", req.AppsTotal))
		return nil
	}

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultAppBranchTimeout,
		RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: defaultAppBranchAttempts},
	})

	processedThisRun := 0
	for len(req.Pending) > 0 && processedThisRun < defaultAppBranchesPerRun {
		appID := req.Pending[0]
		req.Pending = req.Pending[1:]
		processedThisRun++

		progress.CurrentAppID = appID
		resp, err := activities.AwaitEnsureDefaultAppBranch(actCtx, activities.EnsureDefaultAppBranchRequest{AppID: appID})
		if err != nil {
			req.Failed++
			if len(req.FailedAppIDs) < defaultAppBranchMaxFailedIDs {
				req.FailedAppIDs = append(req.FailedAppIDs, appID)
			}
			l.Warn("unable to ensure default app branch", zap.String("app_id", appID), zap.Error(err))
		} else {
			req.InstallsConnected += resp.InstallsConnected
			switch resp.Outcome {
			case activities.DefaultAppBranchOutcomeCreated:
				req.Created++
			case activities.DefaultAppBranchOutcomeClaimed:
				req.Claimed++
			default:
				req.Existing++
			}
		}

		progress.Created = req.Created
		progress.Existing = req.Existing
		progress.Claimed = req.Claimed
		progress.Failed = req.Failed
		progress.FailedAppIDs = req.FailedAppIDs
		progress.InstallsConnected = req.InstallsConnected
		progress.AppsDone = req.Created + req.Existing + req.Claimed + req.Failed
	}

	if len(req.Pending) > 0 {
		l.Info("continuing default app branch backfill",
			zap.Int("apps_done", progress.AppsDone),
			zap.Int("apps_total", req.AppsTotal))
		return workflow.NewContinueAsNewError(ctx, defaultappbranches.WorkflowName, req)
	}

	progress.CurrentAppID = ""
	progress.Done = true
	l.Info("backfilled default app branches",
		zap.Int("created", req.Created),
		zap.Int("existing", req.Existing),
		zap.Int("claimed", req.Claimed),
		zap.Int("failed", req.Failed),
		zap.Int("installs_connected", req.InstallsConnected))
	return nil
}
