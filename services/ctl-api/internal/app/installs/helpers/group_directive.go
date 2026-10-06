package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installgrouprelease"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

type PriorInstallUpdate struct {
	WorkflowID     string
	AppBranchRunID string
}

type GroupDirectiveDecision struct {
	Directive       string
	Reason          string
	WaitingOnRunID  string
	PriorWorkflowID string
}

func (h *Helpers) PriorInFlightInstallUpdate(ctx context.Context, installID, excludeWorkflowID string) (*PriorInstallUpdate, error) {
	var version app.InstallAppConfigVersion
	q := h.db.WithContext(ctx).
		Model(&app.InstallAppConfigVersion{}).
		Joins("JOIN install_workflows ON install_workflows.id = install_app_config_versions.workflow_id").
		Where("install_app_config_versions.install_id = ?", installID).
		Where("install_workflows.plan_only = ?", false).
		Where("install_workflows.finished_at IS NULL").
		Where("install_workflows.status->>'status' NOT IN ?", []string{string(app.StatusError), string(app.StatusCancelled)}).
		Order("install_app_config_versions.created_at DESC")
	if excludeWorkflowID != "" {
		q = q.Where("install_workflows.id <> ?", excludeWorkflowID)
	}
	err := q.First(&version).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get in-flight install update: %w", err)
	}
	if version.WorkflowID == nil || *version.WorkflowID == "" {
		return nil, nil
	}
	prior := &PriorInstallUpdate{WorkflowID: *version.WorkflowID}
	if version.AppBranchRunID != nil {
		prior.AppBranchRunID = *version.AppBranchRunID
	}
	return prior, nil
}

func (h *Helpers) InstallWorkflowTerminal(ctx context.Context, workflowID string) (bool, error) {
	if workflowID == "" {
		return true, nil
	}
	var wf app.Workflow
	err := h.db.WithContext(ctx).First(&wf, "id = ?", workflowID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("unable to get workflow: %w", err)
	}
	if !wf.FinishedAt.IsZero() {
		return true, nil
	}
	switch wf.Status.Status {
	case app.StatusError, app.StatusCancelled, app.StatusSuccess:
		return true, nil
	default:
		return false, nil
	}
}

func (h *Helpers) MarkInstallSupersededForInstall(ctx context.Context, appBranchRunID, installID, supersededByRunID string) error {
	return installgrouprelease.SetInstallStatus(ctx, h.db, appBranchRunID, installID, func(install *app.InstallGroupRunInstall) {
		install.Status = installgrouprelease.StatusSuperseded
		install.SupersededByRunID = supersededByRunID
	})
}

func (h *Helpers) ClassifyInstallGroupDirective(ctx context.Context, installID, _, newAppConfigID, _ string) (*GroupDirectiveDecision, error) {
	decision := &GroupDirectiveDecision{Directive: installgrouprelease.DirectiveAwait}

	var install app.Install
	if err := h.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	diff, err := h.AppBranchConfigDiff(ctx, &install, newAppConfigID)
	if err != nil {
		return nil, err
	}
	if installConfigDiffEmpty(diff) {
		decision.Directive = ""
		return decision, nil
	}

	if diff.StackChanged && !install.SandboxMode.Bool {
		pending, err := h.stackPendingCustomer(ctx, install.ID)
		if err != nil {
			return nil, err
		}
		if pending {
			decision.Directive = installgrouprelease.DirectiveRelease
			decision.Reason = installgrouprelease.ReasonStackPendingCustomer
			return decision, nil
		}
	}

	if !diff.StackChanged && installDiffNeedsRunner(diff) {
		reason, err := h.runnerReleaseReason(ctx, install.ID)
		if err != nil {
			return nil, err
		}
		if reason != "" {
			decision.Directive = installgrouprelease.DirectiveRelease
			decision.Reason = reason
			return decision, nil
		}
	}

	return decision, nil
}

func installConfigDiffEmpty(diff *app.InstallConfigDiff) bool {
	if diff == nil {
		return false
	}
	return !diff.StackChanged && !diff.SandboxChanged && !diff.SandboxBuildChanged &&
		len(diff.Added) == 0 && len(diff.Changed) == 0 && len(diff.Removed) == 0
}

func installDiffNeedsRunner(diff *app.InstallConfigDiff) bool {
	if diff == nil {
		return false
	}
	return diff.SandboxChanged || diff.SandboxBuildChanged || len(diff.Added) > 0 || len(diff.Changed) > 0 || len(diff.Removed) > 0
}

func (h *Helpers) stackPendingCustomer(ctx context.Context, installID string) (bool, error) {
	var count int64
	err := h.db.WithContext(ctx).
		Model(&app.InstallStackVersion{}).
		Where("install_id = ?", installID).
		Where("status->>'status' = ?", app.InstallStackVersionStatusPendingUser).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("unable to check pending stack version: %w", err)
	}
	return count > 0, nil
}

func (h *Helpers) runnerReleaseReason(ctx context.Context, installID string) (string, error) {
	groupIDs := h.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: installID, OwnerType: "installs"})

	var statuses []app.RunnerStatus
	res := h.db.WithContext(ctx).
		Model(&app.Runner{}).
		Scopes(scopes.WithDisableViews).
		Where("runner_group_id IN (?)", groupIDs).
		Order("created_at DESC").
		Limit(1).
		Pluck("status", &statuses)
	if res.Error != nil {
		return "", fmt.Errorf("unable to get install runner status: %w", res.Error)
	}
	if len(statuses) == 0 {
		return "", nil
	}
	switch statuses[0] {
	case app.RunnerStatusOffline:
		return installgrouprelease.ReasonRunnerOffline, nil
	case app.RunnerStatusError:
		return installgrouprelease.ReasonRunnerError, nil
	case app.RunnerStatusDisabled:
		return installgrouprelease.ReasonRunnerDisabled, nil
	default:
		return "", nil
	}
}

func (h *Helpers) RecordInstallGroupDirective(ctx context.Context, appBranchRunID, installID, directive, reason, waitingOnRunID string) error {
	if directive != installgrouprelease.DirectiveRelease {
		return nil
	}
	return installgrouprelease.SetInstallStatus(ctx, h.db, appBranchRunID, installID, func(install *app.InstallGroupRunInstall) {
		install.Status = installgrouprelease.StatusForRelease(reason)
		install.ReleaseReason = reason
		install.WaitingOnRunID = waitingOnRunID
	})
}

func (h *Helpers) RecordReleasedInstallTerminal(ctx context.Context, appBranchRunID, installID, status string) error {
	return installgrouprelease.SetInstallStatus(ctx, h.db, appBranchRunID, installID, func(install *app.InstallGroupRunInstall) {
		switch install.Status {
		case installgrouprelease.StatusPendingCustomer, installgrouprelease.StatusQueued:
			install.Status = status
		}
	})
}
