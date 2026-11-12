package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ReconcileStatus string

const (
	ReconcileStatusCurrent    ReconcileStatus = "current"
	ReconcileStatusBehind     ReconcileStatus = "behind"
	ReconcileStatusInProgress ReconcileStatus = "in_progress"
	ReconcileStatusBlocked    ReconcileStatus = "blocked"
)

type ReconcileBlockerCode string

const (
	ReconcileBlockerNoAppBranch            ReconcileBlockerCode = "no_app_branch"
	ReconcileBlockerNoDeployableRun        ReconcileBlockerCode = "no_deployable_run"
	ReconcileBlockerNoMatchingInstallGroup ReconcileBlockerCode = "no_matching_install_group"
	ReconcileBlockerRolloutInFlight        ReconcileBlockerCode = "rollout_in_flight"
	ReconcileBlockerInstallBusy            ReconcileBlockerCode = "install_busy"
	ReconcileBlockerInstallNotActive       ReconcileBlockerCode = "install_not_active"
	ReconcileBlockerRunnerDisabled         ReconcileBlockerCode = "runner_disabled"
)

type ReconcileBlocker struct {
	Code       ReconcileBlockerCode
	Message    string
	WorkflowID string
}

type ReconcileTarget struct {
	AppBranchID        string
	AppBranchName      string
	AppBranchRunID     string
	AppConfigID        string
	InstallGroupID     string
	InstallGroupName   string
	AppliedAppConfigID string

	Current            bool
	InFlightWorkflowID string

	Diff     *app.InstallConfigDiff
	Blockers []ReconcileBlocker
}

func (t *ReconcileTarget) HasTarget() bool {
	return t.AppConfigID != ""
}

func (t *ReconcileTarget) Status() ReconcileStatus {
	switch {
	case t.Current:
		return ReconcileStatusCurrent
	case t.InFlightWorkflowID != "":
		return ReconcileStatusInProgress
	case len(t.Blockers) > 0:
		return ReconcileStatusBlocked
	default:
		return ReconcileStatusBehind
	}
}

func (h *Helpers) ResolveReconcileTarget(ctx context.Context, install *app.Install) (*ReconcileTarget, error) {
	target := &ReconcileTarget{AppliedAppConfigID: install.DeployedAppConfigID()}

	switch install.LifecyclePhase.Phase {
	case lifecyclephase.Deprovisioning, lifecyclephase.Deprovisioned:
		target.Blockers = append(target.Blockers, ReconcileBlocker{
			Code:    ReconcileBlockerInstallNotActive,
			Message: "The install is deprovisioned or being deprovisioned.",
		})
		return target, nil
	}

	if !install.AppBranchID.Valid || install.AppBranchID.String == "" {
		target.Blockers = append(target.Blockers, ReconcileBlocker{
			Code:    ReconcileBlockerNoAppBranch,
			Message: "The install is not on an app branch.",
		})
		return target, nil
	}
	target.AppBranchID = install.AppBranchID.String
	if install.AppBranch != nil {
		target.AppBranchName = install.AppBranch.Name
	}

	resolved, err := h.ResolveAppBranchRunForInstall(ctx, target.AppBranchID, install)
	switch {
	case errors.Is(err, ErrNoMatchingInstallGroup):
		target.Blockers = append(target.Blockers, ReconcileBlocker{
			Code:    ReconcileBlockerNoMatchingInstallGroup,
			Message: "The install does not match any install group on its app branch.",
		})
		return target, nil
	case err != nil:
		return nil, err
	}
	target.InstallGroupID = resolved.InstallGroupID
	target.InstallGroupName = resolved.InstallGroupName

	if resolved.AppConfigID == "" {
		target.Blockers = append(target.Blockers, ReconcileBlocker{
			Code:    ReconcileBlockerNoDeployableRun,
			Message: "The app branch has no completed run to reconcile onto yet.",
		})
		return target, nil
	}
	target.AppBranchRunID = resolved.AppBranchRunID
	target.AppConfigID = resolved.AppConfigID

	if install.DeployedAppConfigID() == target.AppConfigID {
		target.Current = true
		return target, nil
	}

	inFlight, err := h.inFlightInstallUpdateVersion(ctx, install.ID, "")
	if err != nil {
		return nil, err
	}
	if inFlight != nil && inFlight.WorkflowID != nil {
		if inFlight.NewAppConfigID == target.AppConfigID {
			target.InFlightWorkflowID = *inFlight.WorkflowID
			return target, nil
		}
		target.Blockers = append(target.Blockers, ReconcileBlocker{
			Code:       ReconcileBlockerRolloutInFlight,
			Message:    "Another app branch update is already running on this install.",
			WorkflowID: *inFlight.WorkflowID,
		})
	}

	busy, err := h.busyInstallWorkflowID(ctx, install.ID)
	if err != nil {
		return nil, err
	}
	if busy != "" {
		target.Blockers = append(target.Blockers, ReconcileBlocker{
			Code:       ReconcileBlockerInstallBusy,
			Message:    "Another workflow is running on this install.",
			WorkflowID: busy,
		})
	}

	diff, err := h.AppBranchConfigDiff(ctx, install, target.AppConfigID)
	if err != nil {
		return nil, err
	}
	target.Diff = diff

	metadata := map[string]string{}
	if diff != nil && diff.StackChanged {
		metadata[app.WorkflowMetadataKeyStackChanged] = "true"
	}
	err = h.CheckInstallRunnerForWorkflow(ctx, install.ID, app.WorkflowTypeAppBranchConfigUpdate, metadata)
	switch {
	case errors.Is(err, ErrRunnerDisabled):
		target.Blockers = append(target.Blockers, ReconcileBlocker{
			Code:    ReconcileBlockerRunnerDisabled,
			Message: "The install's runner is disabled. Apply the install stack to re-enable it.",
		})
	case err != nil:
		return nil, fmt.Errorf("unable to check install runner: %w", err)
	}

	return target, nil
}

func (h *Helpers) busyInstallWorkflowID(ctx context.Context, installID string) (string, error) {
	var wf app.Workflow
	err := h.db.WithContext(ctx).
		Where(app.Workflow{OwnerID: installID, OwnerType: "installs"}).
		Where("plan_only = ?", false).
		Where("finished_at IS NULL").
		Where("status->>'status' NOT IN ?", []string{string(app.StatusError), string(app.StatusCancelled), string(app.StatusSuccess)}).
		Where("type NOT IN ?", []app.WorkflowType{
			app.WorkflowTypeAppBranchConfigUpdate,
			app.WorkflowTypeActionWorkflowRun,
			app.WorkflowTypeDriftRun,
			app.WorkflowTypeRunbookRun,
		}).
		Order("created_at DESC").
		Limit(1).
		Find(&wf).Error
	if err != nil {
		return "", fmt.Errorf("unable to check in-flight install workflows: %w", err)
	}
	return wf.ID, nil
}
