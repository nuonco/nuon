package updateinstallgroup

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/installgrouperrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/installgroups"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	statusInProgress = "in-progress"
	statusSuccess    = "success"
	statusError      = "error"
	statusCancelled  = "cancelled"
)

const statusPending = "pending"

type enqueuedInstall struct {
	installID  string
	workflowID string
	cb         callback.Ref
}

type installUpdateFailure struct {
	installID  string
	workflowID string
	detail     string
}

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.NoConfigChanges && !run.Force {
		logger.Info("no config changes, skipping install group update")
		return nil
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	installIDs, groupName, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return err
	}

	if len(installIDs) == 0 {
		logger.Info("no installs in group, skipping")
		return nil
	}

	isPreviewApply := s.PreviewInstallID != "" || s.PreviewLabelSelector != nil

	desiredOnly := false
	if !isPreviewApply {
		enabled, err := activities.AwaitOrgHasFeature(ctx, activities.OrgHasFeatureRequest{
			OrgID:   run.OrgID,
			Feature: string(app.OrgFeatureDesiredConfigRollout),
		})
		if err != nil {
			return fmt.Errorf("unable to check desired config rollout feature: %w", err)
		}
		desiredOnly = enabled && !run.PreviewInstallPlanOnly()
	}
	s.desiredOnly = desiredOnly

	s.writePendingInstallMetadata(ctx, groupName, installIDs)

	releaseInstalls := !s.desiredOnly
	if releaseInstalls {
		if err := s.registerDirectiveHandler(ctx); err != nil {
			return err
		}
	}

	enqueued, err := s.enqueueInstallUpdates(ctx, installIDs, run)
	if err != nil {
		return err
	}

	phase := app.InstallGroupRunPhaseDeploy
	if s.desiredOnly {
		phase = app.InstallGroupRunPhaseDesired
	}
	installEntries := make([]app.InstallGroupRunInstall, 0, len(enqueued))
	for _, e := range enqueued {
		installEntries = append(installEntries, app.InstallGroupRunInstall{
			InstallID:  e.installID,
			WorkflowID: e.workflowID,
			Status:     statusInProgress,
			Phase:      phase,
		})
	}

	var groupRunID string
	if !isPreviewApply {
		groupRunResult, err := activities.AwaitCreateInstallGroupRun(ctx, &activities.CreateInstallGroupRunInput{
			AppBranchRunID:   s.RunID,
			InstallGroupID:   s.InstallGroupID,
			InstallGroupName: groupName,
			TotalInstalls:    len(installIDs),
		})
		if err != nil {
			return fmt.Errorf("unable to create install group run: %w", err)
		}
		groupRunID = groupRunResult.InstallGroupRunID

		_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
			InstallGroupRunID: groupRunID,
			Installs:          installEntries,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: s.inProgressDesc(len(enqueued)),
			},
		})
	}

	s.updateInstallMetadata(ctx, groupName, enqueued, nil)

	var failure *installUpdateFailure
	var awaitErr error
	if releaseInstalls {
		failure, awaitErr = s.awaitInstallUpdatesReleased(ctx, groupName, enqueued, groupRunID, installEntries)
	} else {
		failure, awaitErr = s.awaitInstallUpdates(ctx, groupName, enqueued, groupRunID, installEntries)
	}

	// workflow.Now, not time.Now: wall-clock reads are non-deterministic across
	// replay, so the recorded completion time has to come from workflow time.
	now := workflow.Now(ctx)
	completed, failed, cancelled := countInstallOutcomes(installEntries)
	finalStatus := app.StatusSuccess
	desc := s.progressDesc(completed, failed, cancelled, len(enqueued))
	switch {
	case cancelled > 0:
		finalStatus = app.StatusCancelled
		if failed > 0 {
			desc += fmt.Sprintf(" (%d failed)", failed)
		}
	case failed > 0:
		finalStatus = app.StatusError
	}

	if groupRunID != "" {
		_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
			InstallGroupRunID: groupRunID,
			Installs:          installEntries,
			CompletedInstalls: completed,
			FailedInstalls:    failed,
			CompletedAt:       &now,
			Status: app.CompositeStatus{
				Status:                 finalStatus,
				StatusHumanDescription: desc,
			},
		})
	}

	if failure != nil {
		s.recordInstallUpdateFailure(ctx, failure)
	}

	return awaitErr
}

func countInstallOutcomes(entries []app.InstallGroupRunInstall) (completed, failed, cancelled int) {
	for _, e := range entries {
		switch e.Status {
		case statusSuccess:
			completed++
		case statusError:
			failed++
		case statusCancelled:
			cancelled++
		}
	}
	return
}

func (s *Signal) resolveInstallIDs(ctx workflow.Context) ([]string, string, error) {
	if s.PreviewInstallID != "" || s.PreviewLabelSelector != nil {
		resolved, err := installgroups.ResolvePreviewTarget(ctx, s.AppBranchID, s.PreviewInstallID, s.PreviewLabelSelector)
		if err != nil {
			return nil, "", err
		}
		name := s.SyntheticGroupName
		if name == "" {
			name = resolved.GroupName
		}
		return resolved.InstallIDs, name, nil
	}
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
	if s.desiredOnly {
		return s.enqueueDesiredOnlyUpdates(ctx, installIDs, run)
	}

	logger := workflow.GetLogger(ctx)

	enqueued := make([]enqueuedInstall, 0, len(installIDs))
	for _, installID := range installIDs {
		cb := callback.New(ctx, installID)

		result, err := activities.AwaitCreateInstallAppConfigVersionWorkflow(ctx, &activities.CreateInstallAppConfigVersionWorkflowInput{
			InstallID:      installID,
			NewAppConfigID: run.AppConfigID,
			AppBranchRunID: s.RunID,
			InstallGroupID: s.InstallGroupID,
			PlanOnly:       run.PreviewInstallPlanOnly(),
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

		s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusInProgress, s.statusDesc(false))
	}

	return enqueued, nil
}

// enqueueDesiredOnlyUpdates records each install's desired config and enqueues
// the update-app-config signal. No install workflow is created and nothing is
// deployed; the user reconciles live state to desired from the drift view.
func (s *Signal) enqueueDesiredOnlyUpdates(
	ctx workflow.Context,
	installIDs []string,
	run *app.AppBranchRun,
) ([]enqueuedInstall, error) {
	logger := workflow.GetLogger(ctx)

	enqueued := make([]enqueuedInstall, 0, len(installIDs))
	for _, installID := range installIDs {
		cb := callback.New(ctx, installID)

		result, err := activities.AwaitRecordInstallDesiredConfig(ctx, &activities.RecordInstallDesiredConfigInput{
			InstallID:      installID,
			NewAppConfigID: run.AppConfigID,
			AppBranchRunID: s.RunID,
			InstallGroupID: s.InstallGroupID,
			Callback:       cb,
		})
		if err != nil {
			return nil, fmt.Errorf("install %s: unable to record desired config update: %w", installID, err)
		}

		logger.Info(
			"recorded install desired config update",
			"install_id", installID,
			"install_config_update_id", result.InstallAppConfigVersionID,
		)

		enqueued = append(enqueued, enqueuedInstall{
			installID: installID,
			cb:        cb,
		})

		s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusInProgress, s.statusDesc(false))
	}

	return enqueued, nil
}

func (s *Signal) awaitInstallUpdates(
	ctx workflow.Context,
	groupName string,
	enqueued []enqueuedInstall,
	groupRunID string,
	installEntries []app.InstallGroupRunInstall,
) (*installUpdateFailure, error) {
	logger := workflow.GetLogger(ctx)

	results := make(map[string]string, len(enqueued))
	var errs []error
	var cancelledErr error
	var failure *installUpdateFailure
	noteFailure := func(installID, workflowID, detail string) {
		if failure != nil {
			return
		}
		failure = &installUpdateFailure{
			installID:  installID,
			workflowID: workflowID,
			detail:     detail,
		}
	}
	markCancelled := func(i int, installID, detail string, cause error) {
		results[installID] = statusCancelled
		installEntries[i].Status = statusCancelled
		s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusCancelled, detail)
		if cancelledErr == nil {
			if cause != nil {
				cancelledErr = cause
			} else {
				cancelledErr = temporal.NewNonRetryableApplicationError(
					detail,
					callback.CancelledErrType,
					nil,
				)
			}
		}
	}

	for i, e := range enqueued {
		res, err := callback.AwaitWithTimeout(ctx, e.cb, callback.FallbackAwaitTimeout)
		switch {
		case err != nil && callback.IsCancelled(err):
			markCancelled(i, e.installID, err.Error(), err)

		case err != nil:
			errs = append(errs, fmt.Errorf("install %s %s: %w", e.installID, s.updateTarget(e.workflowID), err))
			noteFailure(e.installID, e.workflowID, err.Error())
			results[e.installID] = statusError
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
			errMsg := fmt.Sprintf("install %s %s: finished as %s", e.installID, s.updateTarget(e.workflowID), status)
			if status == statusCancelled {
				markCancelled(i, e.installID, errMsg, nil)
				break
			}
			errs = append(errs, fmt.Errorf("%s", errMsg))
			noteFailure(e.installID, e.workflowID, errMsg)
			results[e.installID] = statusError
			installEntries[i].Status = statusError
			s.updateInstallAppConfigVersionStatus(ctx, e.installID, app.StatusError, errMsg)

		default:
			results[e.installID] = statusSuccess
			installEntries[i].Status = statusSuccess
			s.updateInstallAppConfigVersionStatus(ctx, e.installID, app.StatusSuccess, s.statusDesc(true))

			logger.Info(
				"install config update completed",
				"install_id", e.installID,
				"workflow_id", e.workflowID,
			)
		}

		completed, failed, cancelled := countInstallOutcomes(installEntries)
		desc := s.progressDesc(completed, failed, cancelled, len(enqueued))

		if groupRunID != "" {
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
		}

		s.updateInstallMetadata(ctx, groupName, enqueued, results)
	}

	if cancelledErr != nil {
		return nil, cancelledErr
	}

	if len(errs) > 0 {
		return failure, fmt.Errorf("update install group had %d errors: %v", len(errs), errs)
	}

	return nil, nil
}

func (s *Signal) recordInstallUpdateFailure(ctx workflow.Context, failure *installUpdateFailure) {
	logger := workflow.GetLogger(ctx)

	name := failure.installID
	install, err := activities.AwaitGetInstallByInstallID(ctx, failure.installID)
	if err != nil {
		logger.Warn("unable to load install for deploy failure", "install_id", failure.installID, "error", err)
	} else if install != nil && install.Name != "" {
		name = install.Name
	}

	if _, err := activities.AwaitSetAppBranchRunInstallUpdateFailedError(ctx, activities.SetAppBranchRunInstallUpdateFailedErrorRequest{
		RunID:       s.RunID,
		InstallID:   failure.installID,
		InstallName: name,
		WorkflowID:  failure.workflowID,
		Detail:      failure.detail,
	}); err != nil {
		logger.Warn("unable to record install update error on app branch run", "run_id", s.RunID, "error", err)
	}

	if s.StepID == "" {
		return
	}

	data, err := compositeerrors.New(
		&installgrouperrors.InstallUpdateFailedError{
			InstallID:   failure.installID,
			InstallName: name,
			WorkflowID:  failure.workflowID,
			Detail:      failure.detail,
		},
		compositeerrors.WithSource("installs", failure.installID),
	)
	if err != nil {
		logger.Warn("unable to build install update composite error", "error", err)
		return
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: data.Message,
			CompositeError:         data,
		},
	}); err != nil {
		logger.Warn("unable to record install update composite error", "step_id", s.StepID, "error", err)
	}
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

func (s *Signal) updateInstallAppConfigVersionStatus(ctx workflow.Context, installID string, status app.Status, desc string) {
	_, _ = activities.AwaitUpdateInstallAppConfigVersionStatus(ctx, &activities.UpdateInstallAppConfigVersionStatusInput{
		AppBranchRunID: s.RunID,
		InstallID:      installID,
		Status:         status,
		StatusDesc:     desc,
	})
}

func (s *Signal) writePendingInstallMetadata(ctx workflow.Context, groupName string, installIDs []string) {
	if s.StepID == "" {
		return
	}

	installs := make([]any, 0, len(installIDs))
	for _, id := range installIDs {
		installs = append(installs, map[string]any{
			"install_id": id,
			"status":     statusPending,
			"phase":      s.installPhase(),
		})
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: s.startingDesc(len(installIDs)),
			Metadata: map[string]any{
				"install_group_name": groupName,
				"total_installs":     len(installIDs),
				"installs":           installs,
			},
		},
	})
}

func (s *Signal) updateInstallMetadata(ctx workflow.Context, groupName string, enqueued []enqueuedInstall, results map[string]string) {
	if s.StepID == "" {
		return
	}

	installs := make([]any, 0, len(enqueued))
	completed := 0
	failed := 0
	cancelled := 0
	for _, e := range enqueued {
		status := statusInProgress
		if results != nil {
			if st, ok := results[e.installID]; ok {
				status = st
				switch st {
				case statusSuccess:
					completed++
				case statusError:
					failed++
				case statusCancelled:
					cancelled++
				}
			}
		}

		installs = append(installs, map[string]any{
			"install_id":  e.installID,
			"workflow_id": e.workflowID,
			"status":      status,
			"phase":       s.installPhase(),
		})
	}

	desc := s.inProgressDesc(len(enqueued))
	if completed > 0 || failed > 0 || cancelled > 0 {
		desc = s.progressDesc(completed, failed, cancelled, len(enqueued))
	}

	metadata := map[string]any{
		"install_group_name": groupName,
		"total_installs":     len(enqueued),
		"installs":           installs,
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: desc,
			Metadata:               metadata,
		},
	})
}

func (s *Signal) installPhase() string {
	if s.desiredOnly {
		return app.InstallGroupRunPhaseDesired
	}
	return app.InstallGroupRunPhaseDeploy
}

func (s *Signal) startingDesc(total int) string {
	if s.desiredOnly {
		return fmt.Sprintf("starting desired config updates for %d installs", total)
	}
	return fmt.Sprintf("starting deploys for %d installs", total)
}

func (s *Signal) inProgressDesc(total int) string {
	if s.desiredOnly {
		return fmt.Sprintf("updating %d installs to the desired app branch run", total)
	}
	return fmt.Sprintf("deploying to %d installs", total)
}

func (s *Signal) progressDesc(completed, failed, cancelled, total int) string {
	outcome := "deployed"
	if s.desiredOnly {
		outcome = "updated to the desired app branch run"
	}
	desc := fmt.Sprintf("%d/%d installs %s", completed, total, outcome)
	if cancelled > 0 {
		desc += fmt.Sprintf(" (%d cancelled)", cancelled)
	}
	if failed > 0 {
		desc += fmt.Sprintf(" (%d failed)", failed)
	}
	return desc
}

func (s *Signal) statusDesc(completed bool) string {
	if s.desiredOnly {
		if completed {
			return "desired config update completed"
		}
		return "desired config update running"
	}
	if completed {
		return "install workflow completed"
	}
	return "install workflow running"
}

func (s *Signal) updateTarget(workflowID string) string {
	if s.desiredOnly {
		return "desired config update"
	}
	return "workflow " + workflowID
}
