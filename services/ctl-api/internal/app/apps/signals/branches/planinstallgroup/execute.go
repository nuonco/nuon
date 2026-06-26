package planinstallgroup

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/workflowstepapprovalrequest"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type installPlanEntry struct {
	InstallID             string                 `json:"install_id"`
	InstallName           string                 `json:"install_name,omitempty"`
	InstallLabels         map[string]string      `json:"install_labels,omitempty"`
	Status                string                 `json:"status"`
	InstallConfigUpdateID string                 `json:"install_config_update_id,omitempty"`
	Diff                  *app.InstallConfigDiff `json:"diff,omitempty"`
	OldAppConfigID        string                 `json:"old_app_config_id,omitempty"`
	NewAppConfigID        string                 `json:"new_app_config_id,omitempty"`
}

type installGroupPlan struct {
	InstallGroup string             `json:"install_group"`
	Installs     []installPlanEntry `json:"installs"`
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
		logger.Info("no installs in group, skipping plan")
		return nil
	}

	entries := make([]installPlanEntry, len(installIDs))
	for i, installID := range installIDs {
		entries[i] = installPlanEntry{
			InstallID: installID,
			Status:    "pending",
		}
	}
	s.updatePlanMetadata(ctx, groupName, entries)

	for i, installID := range installIDs {
		entries[i].Status = "computing"
		s.updatePlanMetadata(ctx, groupName, entries)

		result, err := activities.AwaitCreateInstallConfigUpdate(ctx, &activities.CreateInstallConfigUpdateInput{
			InstallID:      installID,
			NewAppConfigID: run.AppConfigID,
			AppBranchRunID: s.RunID,
			InstallGroupID: s.InstallGroupID,
		})
		if err != nil {
			entries[i].Status = "error"
			s.updatePlanMetadata(ctx, groupName, entries)
			return fmt.Errorf("install %s: unable to create config update: %w", installID, err)
		}

		entries[i].InstallConfigUpdateID = result.InstallConfigUpdateID
		entries[i].Diff = result.Diff
		entries[i].InstallName = result.InstallName
		entries[i].InstallLabels = result.InstallLabels
		entries[i].OldAppConfigID = result.OldAppConfigID
		entries[i].NewAppConfigID = result.NewAppConfigID

		entries[i].Status = "success"
		s.updatePlanMetadata(ctx, groupName, entries)
	}

	if s.StepID == "" {
		logger.Info("no step context, skipping approval dispatch")
		return nil
	}

	plan := installGroupPlan{
		InstallGroup: groupName,
		Installs:     entries,
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("unable to marshal plan: %w", err)
	}

	if err := workflowstepapprovalrequest.Dispatch(ctx, &workflowstepapprovalrequest.Signal{
		InstallID:         installIDs[0],
		InstallWorkflowID: s.FlowID,
		WorkflowStepID:    s.StepID,
		OwnerID:           s.RunID,
		OwnerType:         "app_branch_runs",
		ApprovalType:      app.AppBranchPlanApprovalType,
		Plan:              string(planJSON),
	}); err != nil {
		return fmt.Errorf("unable to dispatch approval request: %w", err)
	}

	logger.Info("install group plan completed with approval request",
		"install_group_id", s.InstallGroupID,
		"install_count", len(installIDs),
	)

	return nil
}

func (s *Signal) resolveInstallIDs(ctx workflow.Context) ([]string, string, error) {
	logger := workflow.GetLogger(ctx)

	group, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return nil, "", fmt.Errorf("unable to get install group: %w", err)
	}

	if group.LabelSelector == nil {
		logger.Info("planning install group",
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

	logger.Info("planning install group",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(resolved.InstallIDs),
		"resolved_via", "label_selector",
	)

	return resolved.InstallIDs, group.Name, nil
}

func (s *Signal) updatePlanMetadata(ctx workflow.Context, groupName string, entries []installPlanEntry) {
	if s.StepID == "" {
		return
	}

	installs := make([]any, 0, len(entries))
	completed := 0
	for _, e := range entries {
		entry := map[string]any{
			"install_id": e.InstallID,
			"status":     e.Status,
		}
		if e.InstallName != "" {
			entry["install_name"] = e.InstallName
		}
		if e.InstallConfigUpdateID != "" {
			entry["install_config_update_id"] = e.InstallConfigUpdateID
		}
		if e.Diff != nil {
			entry["added"] = len(e.Diff.Added)
			entry["changed"] = len(e.Diff.Changed)
			entry["removed"] = len(e.Diff.Removed)
			entry["unchanged"] = len(e.Diff.Unchanged)
			entry["sandbox_changed"] = e.Diff.SandboxChanged
			entry["stack_changed"] = e.Diff.StackChanged
		}
		installs = append(installs, entry)
		if e.Status == "success" {
			completed++
		}
	}

	desc := fmt.Sprintf("computing plan for %d installs", len(entries))
	if completed == len(entries) {
		desc = fmt.Sprintf("plan computed for %d installs", len(entries))
	} else if completed > 0 {
		desc = fmt.Sprintf("computing plan: %d/%d installs", completed, len(entries))
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: desc,
			Metadata: map[string]any{
				"install_group_name": groupName,
				"total_installs":     len(entries),
				"installs":           installs,
			},
		},
	})
}
