package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ResolveAllInstallsForBranch returns the app's installs that no other branch
// owns, which is what an install group with AllInstalls set targets.
//
// Ownership is checked two ways because the data model records it two ways: an
// install pinned to a group or labelled into one carries app_branch_id, but an
// install created with labels that a selector happens to match never gets it
// written. Missing the second case would let the default branch deploy over a
// VCS branch's installs.
func (h *Helpers) ResolveAllInstallsForBranch(ctx context.Context, appID, branchID string) ([]string, error) {
	var installs []app.Install
	if err := h.db.WithContext(ctx).
		Where(app.Install{AppID: appID}).
		Where("app_branch_id IS NULL OR app_branch_id = ?", branchID).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to load installs: %w", err)
	}

	claimed, err := h.otherBranchSelectors(ctx, appID, branchID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(installs))
	for _, install := range installs {
		if matchesAny(claimed, install.Labels) {
			continue
		}
		ids = append(ids, install.ID)
	}

	return ids, nil
}

// otherBranchSelectors collects the label selectors on every other branch's
// latest config.
func (h *Helpers) otherBranchSelectors(ctx context.Context, appID, branchID string) ([]*labels.Selector, error) {
	var branches []app.AppBranch
	if err := h.db.WithContext(ctx).
		Where(app.AppBranch{AppID: appID}).
		Where("id != ?", branchID).
		Find(&branches).Error; err != nil {
		return nil, fmt.Errorf("unable to load app branches: %w", err)
	}

	var selectors []*labels.Selector
	for _, branch := range branches {
		var latest app.AppBranchConfig
		err := h.db.WithContext(ctx).
			Where(app.AppBranchConfig{AppBranchID: branch.ID}).
			Order("created_at DESC").
			First(&latest).Error
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("unable to load latest config for branch %s: %w", branch.ID, err)
		}

		var groups []app.AppBranchInstallGroup
		if err := h.db.WithContext(ctx).
			Where(app.AppBranchInstallGroup{AppBranchConfigID: latest.ID}).
			Find(&groups).Error; err != nil {
			return nil, fmt.Errorf("unable to load install groups for config %s: %w", latest.ID, err)
		}

		for i := range groups {
			if groups[i].LabelSelector != nil {
				selectors = append(selectors, groups[i].LabelSelector)
			}
		}
	}

	return selectors, nil
}

func matchesAny(selectors []*labels.Selector, lbls labels.Labels) bool {
	for _, selector := range selectors {
		if selector.Matches(lbls) {
			return true
		}
	}
	return false
}
