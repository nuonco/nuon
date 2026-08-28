package updateinstallgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/installgroups"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updateappconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	statusInProgress = "in-progress"
	statusSuccess    = "success"
	statusError      = "error"
)

// installVersionStatusVersion gates the per-install config-version status
// writes added by the diffing engine; in-flight histories never scheduled
// those activities between the enqueue and await commands.
// todo(sk): cleanup after terminating old workflows
const installVersionStatusVersion = "install-app-config-version-status-v1"

type enqueuedInstall struct {
	installID  string
	workflowID string
	cb         callback.Ref
}

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	installIDs, groupName, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return err
	}

	if len(installIDs) == 0 {
		logger.Info("no installs in group, skipping")
		return nil
	}

	groupRunResult, err := activities.AwaitCreateInstallGroupRun(ctx, &activities.CreateInstallGroupRunInput{
		AppBranchRunID:   s.RunID,
		InstallGroupID:   s.InstallGroupID,
		InstallGroupName: groupName,
		TotalInstalls:    len(installIDs),
	})
	if err != nil {
		return fmt.Errorf("unable to create install group run: %w", err)
	}

	enqueued, err := s.enqueueInstallUpdates(ctx, installIDs, run)
	if err != nil {
		return err
	}

	installEntries := make([]app.InstallGroupRunInstall, 0, len(enqueued))
	for _, e := range enqueued {
		installEntries = append(installEntries, app.InstallGroupRunInstall{
			InstallID:  e.installID,
			WorkflowID: e.workflowID,
			Status:     statusInProgress,
			Phase:      app.InstallGroupRunPhaseDeploy,
		})
	}

	_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
		InstallGroupRunID: groupRunResult.InstallGroupRunID,
		Installs:          installEntries,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: fmt.Sprintf("deploying to %d installs", len(enqueued)),
		},
	})

	s.updateInstallMetadata(ctx, enqueued, nil)

	completed, failed, awaitErr := s.awaitInstallUpdates(ctx, enqueued, groupRunResult.InstallGroupRunID, installEntries)

	// workflow.Now, not time.Now: wall-clock reads are non-deterministic across
	// replay, so the recorded completion time has to come from workflow time.
	now := workflow.Now(ctx)
	finalStatus := app.StatusSuccess
	desc := fmt.Sprintf("%d/%d installs deployed", completed, len(enqueued))
	if failed > 0 {
		finalStatus = app.StatusError
		desc += fmt.Sprintf(" (%d failed)", failed)
	}

	_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
		InstallGroupRunID: groupRunResult.InstallGroupRunID,
		Installs:          installEntries,
		CompletedInstalls: completed,
		FailedInstalls:    failed,
		CompletedAt:       &now,
		Status: app.CompositeStatus{
			Status:                 finalStatus,
			StatusHumanDescription: desc,
		},
	})

	return awaitErr
}

func (s *Signal) resolveInstallIDs(ctx workflow.Context) ([]string, string, error) {
	resolved, err := installgroups.Resolve(ctx, s.InstallGroupID, s.AppBranchID)
	if err != nil {
		return nil, "", err
	}
	return resolved.InstallIDs, resolved.GroupName, nil
}

func (s *Signal) enqueueInstallUpdates(
	ctx workflow.Context,
	installIDs []string,
	run *app.AppBranchRun,
) ([]enqueuedInstall, error) {
	logger := workflow.GetLogger(ctx)

	enqueued := make([]enqueuedInstall, 0, len(installIDs))
	for _, installID := range installIDs {
		cb := callback.New(ctx, installID)

		result, err := activities.AwaitCreateInstallAppConfigVersionWorkflow(ctx, &activities.CreateInstallAppConfigVersionWorkflowInput{
			InstallID:      installID,
			NewAppConfigID: run.AppConfigID,
			AppBranchRunID: s.RunID,
			InstallGroupID: s.InstallGroupID,
			PlanOnly:       run.PlanOnly,
			Callback:       cb,
		})
		if err != nil {
			return nil, fmt.Errorf("install %s: unable to create config update workflow: %w", installID, err)
		}

		logger.Info(
			"enqueued install config update",
			"install_id", installID,
			"workflow_id", result.WorkflowID,
			"install_config_update_id", result.InstallAppConfigVersionID,
		)

		s.childWorkflowIDs = append(s.childWorkflowIDs, result.WorkflowID)
		enqueued = append(enqueued, enqueuedInstall{
			installID:  installID,
			workflowID: result.WorkflowID,
			cb:         cb,
		})

		s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusInProgress, "install workflow running")
	}

	return enqueued, nil
}

func (s *Signal) awaitInstallUpdates(
	ctx workflow.Context,
	enqueued []enqueuedInstall,
	groupRunID string,
	installEntries []app.InstallGroupRunInstall,
) (int, int, error) {
	logger := workflow.GetLogger(ctx)

	completed := 0
	failed := 0
	results := make(map[string]string, len(enqueued))
	var errs []error

	for i, e := range enqueued {
		res, err := callback.AwaitWithTimeout(ctx, e.cb, callback.FallbackAwaitTimeout)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("install %s workflow %s: %w", e.installID, e.workflowID, err))
			results[e.installID] = statusError
			failed++
			installEntries[i].Status = statusError
			s.updateInstallAppConfigVersionStatus(ctx, e.installID, app.StatusError, err.Error())

		// Only an "error" result comes back as a Go error, so a cancelled or expired
		// deploy arrives here as a clean return and would otherwise be counted as a
		// successful install.
		case res == nil || res.Status != statusSuccess:
			status := "unknown"
			if res != nil && res.Status != "" {
				status = res.Status
			}
			errMsg := fmt.Sprintf("install %s workflow %s: finished as %s", e.installID, e.workflowID, status)
			errs = append(errs, fmt.Errorf("%s", errMsg))
			results[e.installID] = statusError
			failed++
			installEntries[i].Status = statusError
			s.updateInstallAppConfigVersionStatus(ctx, e.installID, app.StatusError, errMsg)

		default:
			results[e.installID] = statusSuccess
			completed++
			installEntries[i].Status = statusSuccess
			s.updateInstallAppConfigVersionStatus(ctx, e.installID, app.StatusSuccess, "install workflow completed")

			logger.Info(
				"install config update completed",
				"install_id", e.installID,
				"workflow_id", e.workflowID,
			)
		}

		desc := fmt.Sprintf("%d/%d installs deployed", completed, len(enqueued))
		if failed > 0 {
			desc += fmt.Sprintf(" (%d failed)", failed)
		}

		_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
			InstallGroupRunID: groupRunID,
			Installs:          installEntries,
			CompletedInstalls: completed,
			FailedInstalls:    failed,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: desc,
			},
		})

		s.updateInstallMetadata(ctx, enqueued, results)
	}

	if len(errs) > 0 {
		return completed, failed, fmt.Errorf("update install group had %d errors: %v", len(errs), errs)
	}

	return completed, failed, nil
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info(
		"cancelling install group update",
		"install_group_id", s.InstallGroupID,
		"child_workflow_count", len(s.childWorkflowIDs),
	)

	for _, wfID := range s.childWorkflowIDs {
		if err := activities.AwaitCancelInstallWorkflow(ctx, &activities.CancelInstallWorkflowInput{
			WorkflowID: wfID,
		}); err != nil {
			logger.Warn(
				"failed to cancel child workflow",
				"workflow_id", wfID,
				"error", err,
			)
		}
	}

	return nil
}

func (s *Signal) recordAppConfigVersions(
	ctx workflow.Context,
	enqueued []enqueuedInstall,
	installEntries []app.InstallGroupRunInstall,
	run *app.AppBranchRun,
) {
	logger := workflow.GetLogger(ctx)

	for i, e := range enqueued {
		if installEntries[i].Status != "success" {
			continue
		}

		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   e.installID,
			OwnerType: "installs",
			QueueName: "install-signals",
			Signal: &updateappconfig.Signal{
				InstallID:      e.installID,
				NewAppConfigID: run.AppConfigID,
				AppBranchRunID: s.RunID,
				InstallGroupID: s.InstallGroupID,
				TriggeredBy:    "app-branch",
				Metadata:       map[string]string{"source": "app-branch"},
			},
		}); err != nil {
			logger.Warn(
				"unable to enqueue update-app-config signal",
				"install_id", e.installID,
				"error", err,
			)
		}
	}
}

func (s *Signal) updateInstallAppConfigVersionStatus(ctx workflow.Context, installID string, status app.Status, desc string) {
	if workflow.GetVersion(ctx, installVersionStatusVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return
	}
	_, _ = activities.AwaitUpdateInstallAppConfigVersionStatus(ctx, &activities.UpdateInstallAppConfigVersionStatusInput{
		AppBranchRunID: s.RunID,
		InstallID:      installID,
		Status:         status,
		StatusDesc:     desc,
	})
}

func (s *Signal) updateInstallMetadata(ctx workflow.Context, enqueued []enqueuedInstall, results map[string]string) {
	if s.StepID == "" {
		return
	}

	installs := make([]any, 0, len(enqueued))
	for _, e := range enqueued {
		status := statusInProgress
		if results != nil {
			if s, ok := results[e.installID]; ok {
				status = s
			}
		}

		installs = append(installs, map[string]any{
			"install_id":  e.installID,
			"workflow_id": e.workflowID,
			"status":      status,
		})
	}

	completed := 0
	failed := 0
	if results != nil {
		for _, s := range results {
			switch s {
			case statusSuccess:
				completed++
			case statusError:
				failed++
			}
		}
	}

	desc := fmt.Sprintf("deploying to %d installs", len(enqueued))
	if completed > 0 || failed > 0 {
		desc = fmt.Sprintf("%d/%d installs deployed", completed, len(enqueued))
		if failed > 0 {
			desc += fmt.Sprintf(" (%d failed)", failed)
		}
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: desc,
			Metadata: map[string]any{
				"installs": installs,
			},
		},
	})
}
