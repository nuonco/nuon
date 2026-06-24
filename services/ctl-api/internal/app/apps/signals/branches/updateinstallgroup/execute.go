package updateinstallgroup

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

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
			Status:     "in-progress",
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

	now := time.Now()
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
	logger := workflow.GetLogger(ctx)

	group, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return nil, "", fmt.Errorf("unable to get install group: %w", err)
	}

	if group.LabelSelector == nil {
		logger.Info("updating install group",
			"install_group_id", group.ID,
			"install_group_name", group.Name,
			"install_count", len(group.InstallIDs),
		)
		return group.InstallIDs, group.Name, nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return nil, "", fmt.Errorf("unable to get app branch for label resolution: %w", err)
	}

	resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
		AppID:    branch.AppID,
		GroupID:  group.ID,
		Selector: group.LabelSelector,
	})
	if err != nil {
		return nil, "", fmt.Errorf("unable to resolve install group labels: %w", err)
	}

	logger.Info("updating install group",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(resolved.InstallIDs),
		"resolved_via", "label_selector",
	)

	return resolved.InstallIDs, group.Name, nil
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

		result, err := activities.AwaitCreateInstallConfigUpdateWorkflow(ctx, &activities.CreateInstallConfigUpdateWorkflowInput{
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

		logger.Info("enqueued install config update",
			"install_id", installID,
			"workflow_id", result.WorkflowID,
			"install_config_update_id", result.InstallConfigUpdateID,
		)

		s.childWorkflowIDs = append(s.childWorkflowIDs, result.WorkflowID)
		enqueued = append(enqueued, enqueuedInstall{
			installID:  installID,
			workflowID: result.WorkflowID,
			cb:         cb,
		})
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
		if _, err := callback.AwaitWithTimeout(ctx, e.cb, callback.FallbackAwaitTimeout); err != nil {
			errs = append(errs, fmt.Errorf("install %s workflow %s: %w", e.installID, e.workflowID, err))
			results[e.installID] = "error"
			failed++
			installEntries[i].Status = "error"
		} else {
			results[e.installID] = "success"
			completed++
			installEntries[i].Status = "success"

			logger.Info("install config update completed",
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
	logger.Info("cancelling install group update",
		"install_group_id", s.InstallGroupID,
		"child_workflow_count", len(s.childWorkflowIDs),
	)

	for _, wfID := range s.childWorkflowIDs {
		if err := activities.AwaitCancelInstallWorkflow(ctx, &activities.CancelInstallWorkflowInput{
			WorkflowID: wfID,
		}); err != nil {
			logger.Warn("failed to cancel child workflow",
				"workflow_id", wfID,
				"error", err,
			)
		}
	}

	return nil
}

func (s *Signal) updateInstallMetadata(ctx workflow.Context, enqueued []enqueuedInstall, results map[string]string) {
	if s.StepID == "" {
		return
	}

	installs := make([]any, 0, len(enqueued))
	for _, e := range enqueued {
		status := "in-progress"
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
			case "success":
				completed++
			case "error":
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
