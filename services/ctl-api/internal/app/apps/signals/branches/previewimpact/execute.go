package previewimpact

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/installgroups"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	groups, err := s.computeImpact(ctx, l, run.AppConfigID)
	if err != nil {
		return err
	}

	s.updateStepMetadata(ctx, groups)
	s.updatePRComment(ctx, l, run, groups)

	return nil
}

func (s *Signal) computeImpact(ctx workflow.Context, l log.Logger, newAppConfigID string) ([]activities.InstallGroupImpact, error) {
	installGroups, err := activities.AwaitGetInstallGroupsByConfigID(ctx, s.AppBranchConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch install groups: %w", err)
	}

	impact := make([]activities.InstallGroupImpact, 0, len(installGroups))
	for _, group := range installGroups {
		resolved, err := installgroups.Resolve(ctx, group.ID, s.AppBranchID)
		if err != nil {
			return nil, fmt.Errorf("install group %s: %w", group.Name, err)
		}
		if len(resolved.InstallIDs) == 0 {
			continue
		}

		entry := activities.InstallGroupImpact{
			GroupName: resolved.GroupName,
			Installs:  make([]activities.InstallImpact, 0, len(resolved.InstallIDs)),
		}

		for _, installID := range resolved.InstallIDs {
			install, err := activities.AwaitGetInstallByInstallID(ctx, installID)
			if err != nil {
				return nil, fmt.Errorf("install %s: unable to get install: %w", installID, err)
			}

			diffResult, err := activities.AwaitComputeInstallConfigDiff(ctx, &activities.ComputeInstallConfigDiffInput{
				OldAppConfigID: install.AppConfigID,
				NewAppConfigID: newAppConfigID,
			})
			if err != nil {
				return nil, fmt.Errorf("install %s: unable to compute config diff: %w", installID, err)
			}

			name := install.Name
			if name == "" {
				name = installID
			}
			entry.Installs = append(entry.Installs, activities.InstallImpact{
				InstallID:      installID,
				InstallName:    name,
				Added:          len(diffResult.Diff.Added),
				Changed:        len(diffResult.Diff.Changed),
				Removed:        len(diffResult.Diff.Removed),
				Unchanged:      len(diffResult.Diff.Unchanged),
				SandboxChanged: diffResult.Diff.SandboxChanged,
				StackChanged:   diffResult.Diff.StackChanged,
			})
		}

		impact = append(impact, entry)
	}

	l.Info("computed preview impact", "run_id", s.RunID, "install_group_count", len(impact))
	return impact, nil
}

func (s *Signal) updateStepMetadata(ctx workflow.Context, groups []activities.InstallGroupImpact) {
	if s.StepID == "" {
		return
	}

	total := 0
	groupList := make([]any, 0, len(groups))
	for _, g := range groups {
		installs := make([]any, 0, len(g.Installs))
		for _, i := range g.Installs {
			installs = append(installs, map[string]any{
				"install_id":      i.InstallID,
				"install_name":    i.InstallName,
				"added":           i.Added,
				"changed":         i.Changed,
				"removed":         i.Removed,
				"unchanged":       i.Unchanged,
				"sandbox_changed": i.SandboxChanged,
				"stack_changed":   i.StackChanged,
			})
		}
		total += len(g.Installs)
		groupList = append(groupList, map[string]any{
			"install_group_name": g.GroupName,
			"installs":           installs,
		})
	}

	desc := fmt.Sprintf("no installs affected by this config")
	if total > 0 {
		desc = fmt.Sprintf("previewed impact on %d installs", total)
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: desc,
			Metadata: map[string]any{
				"total_installs": total,
				"install_groups": groupList,
			},
		},
	})
}

func (s *Signal) updatePRComment(ctx workflow.Context, l log.Logger, run *app.AppBranchRun, groups []activities.InstallGroupImpact) {
	if run.PRNumber == nil {
		return
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		l.Warn("unable to get branch for preview comment", "error", err)
		return
	}

	var vcsConfigID string
	if len(branch.Configs) > 0 {
		if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		}
	}
	if vcsConfigID == "" {
		return
	}

	var diff *activities.ComputeAppConfigDiffOutput
	if run.AppConfigID != "" {
		var oldConfigID string
		baseline, baselineErr := activities.AwaitResolvePreviewBaselineAppConfig(ctx, &activities.ResolvePreviewBaselineAppConfigInput{
			RunID:       s.RunID,
			AppBranchID: s.AppBranchID,
		})
		if baselineErr == nil && baseline.AppConfigID != "" {
			oldConfigID = baseline.AppConfigID
		}

		computed, diffErr := activities.AwaitComputeAppConfigDiff(ctx, &activities.ComputeAppConfigDiffInput{
			AppID:       branch.AppID,
			NewConfigID: run.AppConfigID,
			OldConfigID: oldConfigID,
		})
		if diffErr == nil {
			diff = computed
		}
	}

	body := activities.BuildPRCommentBody(&activities.PRCommentParams{
		AppName:       branch.Name,
		RunID:         s.RunID,
		Status:        activities.PRCommentStatusSuccess,
		Diff:          diff,
		InstallImpact: groups,
	})

	if _, err := activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
		ExistingCommentID: run.GithubCommentID,
		Body:              body,
	}); err != nil {
		l.Warn("unable to update preview PR comment", "error", err)
	}
}
