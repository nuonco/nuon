package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ResolveAppBranchApprovalOption returns approve-all only when every install the
// branch config targets is itself set to approve-all. A run has one approval
// option but installs each carry their own, so a single install that prompts
// holds the whole run at its plan step.
func (h *Helpers) ResolveAppBranchApprovalOption(ctx context.Context, appID, branchID, branchConfigID string) (app.InstallApprovalOption, error) {
	installIDs, err := h.resolveAppBranchConfigInstallIDs(ctx, appID, branchID, branchConfigID)
	if err != nil {
		return app.InstallApprovalOptionPrompt, err
	}
	if len(installIDs) == 0 {
		return app.InstallApprovalOptionPrompt, nil
	}

	type installApproval struct {
		InstallID      string
		ApprovalOption app.InstallApprovalOption
	}

	var approvals []installApproval
	if err := h.db.WithContext(ctx).
		Raw(`SELECT DISTINCT ON (install_id) install_id, approval_option
			FROM install_configs
			WHERE install_id IN ? AND deleted_at = 0
			ORDER BY install_id, created_at DESC`, installIDs).
		Scan(&approvals).Error; err != nil {
		return app.InstallApprovalOptionPrompt, fmt.Errorf("unable to load install approval options: %w", err)
	}

	if len(approvals) != len(installIDs) {
		return app.InstallApprovalOptionPrompt, nil
	}

	for _, approval := range approvals {
		if approval.ApprovalOption != app.InstallApprovalOptionApproveAll {
			return app.InstallApprovalOptionPrompt, nil
		}
	}

	return app.InstallApprovalOptionApproveAll, nil
}

// resolveAppBranchConfigInstallIDs mirrors what the run's install group steps
// resolve at execution time, so the approval option is derived from the same set
// of installs the run will actually touch.
func (h *Helpers) resolveAppBranchConfigInstallIDs(ctx context.Context, appID, branchID, branchConfigID string) ([]string, error) {
	var groups []app.AppBranchInstallGroup
	if err := h.db.WithContext(ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: branchConfigID}).
		Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("unable to load install groups: %w", err)
	}

	seen := make(map[string]struct{})
	var ids []string
	add := func(installIDs []string) {
		for _, id := range installIDs {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	for _, group := range groups {
		switch {
		case group.AllInstalls:
			installIDs, err := h.ResolveAllInstallsForBranch(ctx, appID, branchID)
			if err != nil {
				return nil, err
			}
			add(installIDs)
		case group.LabelSelector != nil:
			var installs []app.Install
			if err := h.db.WithContext(ctx).
				Where(app.Install{AppID: appID}).
				Scopes(labels.WithLabels("labels", group.LabelSelector.MatchLabels)).
				Find(&installs).Error; err != nil {
				return nil, fmt.Errorf("unable to resolve install group labels: %w", err)
			}
			for _, install := range installs {
				add([]string{install.ID})
			}
		default:
			add(group.InstallIDs)
		}
	}

	return ids, nil
}
