package planinstallgroup

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/installgroups"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/workflowstepapprovalrequest"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type installPlanEntry struct {
	InstallID      string                 `json:"install_id"`
	InstallName    string                 `json:"install_name,omitempty"`
	InstallLabels  map[string]string      `json:"install_labels,omitempty"`
	Status         string                 `json:"status"`
	Diff           *app.InstallConfigDiff `json:"diff,omitempty"`
	OldAppConfigID string                 `json:"old_app_config_id,omitempty"`
	NewAppConfigID string                 `json:"new_app_config_id,omitempty"`
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

		install, err := activities.AwaitGetInstallByInstallID(ctx, installID)
		if err != nil {
			entries[i].Status = "error"
			s.updatePlanMetadata(ctx, groupName, entries)
			return fmt.Errorf("install %s: unable to get install: %w", installID, err)
		}

		diffResult, err := activities.AwaitComputeInstallConfigDiff(ctx, &activities.ComputeInstallConfigDiffInput{
			OldAppConfigID: install.AppConfigID,
			NewAppConfigID: run.AppConfigID,
		})
		if err != nil {
			entries[i].Status = "error"
			s.updatePlanMetadata(ctx, groupName, entries)
			return fmt.Errorf("install %s: unable to compute config diff: %w", installID, err)
		}

		entries[i].Diff = diffResult.Diff
		entries[i].InstallName = install.Name
		entries[i].InstallLabels = install.Labels
		entries[i].OldAppConfigID = install.AppConfigID
		entries[i].NewAppConfigID = run.AppConfigID

		entries[i].Status = "success"
		s.updatePlanMetadata(ctx, groupName, entries)
	}

	if s.StepID == "" {
		logger.Info("no step context, skipping approval dispatch")
		return nil
	}

	if run.Preview != nil && run.Preview.Mode == app.AppBranchRunPreviewModePlanOnly {
		logger.Info("preview plan-only run, skipping approval dispatch")
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
	previewInstallID := s.PreviewInstallID
	if previewInstallID == "" {
		run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
		if err != nil {
			return nil, "", fmt.Errorf("unable to get app branch run: %w", err)
		}
		if run.Preview != nil {
			previewInstallID = run.Preview.InstallID
		}
	}
	if previewInstallID != "" {
		name := s.SyntheticGroupName
		if name == "" {
			name = "preview"
		}
		return []string{previewInstallID}, name, nil
	}
	resolved, err := installgroups.Resolve(ctx, s.InstallGroupID, s.AppBranchID)
	if err != nil {
		return nil, "", err
	}
	return resolved.InstallIDs, resolved.GroupName, nil
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
		if len(e.InstallLabels) > 0 {
			entry["install_labels"] = e.InstallLabels
		}
		if e.OldAppConfigID != "" {
			entry["old_app_config_id"] = e.OldAppConfigID
		}
		if e.NewAppConfigID != "" {
			entry["new_app_config_id"] = e.NewAppConfigID
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
