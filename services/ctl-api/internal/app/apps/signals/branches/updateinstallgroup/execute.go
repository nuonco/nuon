package updateinstallgroup

import (
	"fmt"

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

	installIDs, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return err
	}

	if len(installIDs) == 0 {
		logger.Info("no installs in group, skipping")
		return nil
	}

	// Create and enqueue all install config update workflows.
	enqueued, err := s.enqueueInstallUpdates(ctx, installIDs, run)
	if err != nil {
		return err
	}

	s.updateInstallMetadata(ctx, enqueued, nil)

	// Wait for all install workflows to complete.
	return s.awaitInstallUpdates(ctx, enqueued)
}

// resolveInstallIDs returns the install IDs for this group, resolving via
// label selector if configured.
func (s *Signal) resolveInstallIDs(ctx workflow.Context) ([]string, error) {
	logger := workflow.GetLogger(ctx)

	group, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install group: %w", err)
	}

	if group.LabelSelector == nil {
		logger.Info("updating install group",
			"install_group_id", group.ID,
			"install_group_name", group.Name,
			"install_count", len(group.InstallIDs),
		)
		return group.InstallIDs, nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branch for label resolution: %w", err)
	}

	resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
		AppID:    branch.AppID,
		GroupID:  group.ID,
		Selector: group.LabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to resolve install group labels: %w", err)
	}

	logger.Info("updating install group",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(resolved.InstallIDs),
		"resolved_via", "label_selector",
	)

	return resolved.InstallIDs, nil
}

// enqueueInstallUpdates creates an install config update workflow for each
// install and enqueues it for execution. All workflows are created and
// enqueued before any awaiting begins.
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

		enqueued = append(enqueued, enqueuedInstall{
			installID:  installID,
			workflowID: result.WorkflowID,
			cb:         cb,
		})
	}

	return enqueued, nil
}

// awaitInstallUpdates waits for all enqueued install workflows to complete.
func (s *Signal) awaitInstallUpdates(ctx workflow.Context, enqueued []enqueuedInstall) error {
	logger := workflow.GetLogger(ctx)

	results := make(map[string]string, len(enqueued))
	var errs []error
	for _, e := range enqueued {
		if _, err := callback.AwaitWithTimeout(ctx, e.cb, callback.FallbackAwaitTimeout); err != nil {
			errs = append(errs, fmt.Errorf("install %s workflow %s: %w", e.installID, e.workflowID, err))
			results[e.installID] = "error"
			s.updateInstallMetadata(ctx, enqueued, results)
			continue
		}

		results[e.installID] = "success"
		s.updateInstallMetadata(ctx, enqueued, results)

		logger.Info("install config update completed",
			"install_id", e.installID,
			"workflow_id", e.workflowID,
		)
	}

	if len(errs) > 0 {
		return fmt.Errorf("update install group had %d errors: %v", len(errs), errs)
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
