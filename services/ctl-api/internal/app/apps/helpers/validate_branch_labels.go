package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type BranchForLabels struct {
	Branch *app.AppBranch
	Group  *app.AppBranchInstallGroup
}

// LatestConfigInstallGroups returns the install groups on a branch's newest
// config, or nothing when the branch has no config yet.
func (h *Helpers) LatestConfigInstallGroups(ctx context.Context, branchID string) ([]app.AppBranchInstallGroup, error) {
	var latestConfig app.AppBranchConfig
	if err := h.db.WithContext(ctx).
		Where(app.AppBranchConfig{AppBranchID: branchID}).
		Order("created_at DESC").
		First(&latestConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest config for branch %s: %w", branchID, err)
	}

	var groups []app.AppBranchInstallGroup
	if err := h.db.WithContext(ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: latestConfig.ID}).
		Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("unable to get install groups for config %s: %w", latestConfig.ID, err)
	}

	return groups, nil
}

// InstallOwnedByOtherBranch reports whether an install is held by a branch other
// than claimingBranchID with a claim strong enough to refuse a handover.
//
// A branch that picked the install up through an all-installs group is a weak
// owner: it claimed whatever was unclaimed rather than this install in
// particular, so a selector or an explicit ID naming the install takes it over.
// Run-time resolution already works this way, since ResolveAllInstallsForBranch
// drops installs another branch's selector matches.
func (h *Helpers) InstallOwnedByOtherBranch(ctx context.Context, install *app.Install, claimingBranchID string) (bool, error) {
	return h.installOwnedByOtherBranch(ctx, install, claimingBranchID, newOwnerGroupCache())
}

// ownerGroupCache keeps a caller that walks many installs from re-reading the
// same owning branch's config for each one.
type ownerGroupCache map[string][]app.AppBranchInstallGroup

func newOwnerGroupCache() ownerGroupCache {
	return ownerGroupCache{}
}

func (h *Helpers) installOwnedByOtherBranch(
	ctx context.Context,
	install *app.Install,
	claimingBranchID string,
	cache ownerGroupCache,
) (bool, error) {
	ownerID := install.AppBranchID.String
	if !install.AppBranchID.Valid || ownerID == "" || ownerID == claimingBranchID {
		return false, nil
	}

	groups, ok := cache[ownerID]
	if !ok {
		loaded, err := h.LatestConfigInstallGroups(ctx, ownerID)
		if err != nil {
			return false, err
		}
		groups = loaded
		cache[ownerID] = groups
	}

	return !claimIsAllInstallsOnly(groups, install), nil
}

// claimIsAllInstallsOnly reports whether the owning branch holds the install
// only because one of its groups takes everything unclaimed. An owner with no
// all-installs group holds it for some reason this cannot see, so it keeps it.
func claimIsAllInstallsOnly(groups []app.AppBranchInstallGroup, install *app.Install) bool {
	allInstalls := false
	for i := range groups {
		group := &groups[i]
		for _, id := range group.InstallIDs {
			if id == install.ID {
				return false
			}
		}
		if group.LabelSelector != nil && group.LabelSelector.Matches(install.Labels) {
			return false
		}
		if group.AllInstalls {
			allInstalls = true
		}
	}

	return allInstalls
}

// FindBranchesMatchingLabels returns all app branches whose latest config
// has an install group with a label selector that matches the given labels.
func (h *Helpers) FindBranchesMatchingLabels(ctx context.Context, appID string, lbls labels.Labels) ([]BranchForLabels, error) {
	var branches []app.AppBranch
	if err := h.db.WithContext(ctx).
		Where(app.AppBranch{AppID: appID}).
		Find(&branches).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branches: %w", err)
	}

	var results []BranchForLabels
	for i := range branches {
		branch := &branches[i]
		groups, err := h.LatestConfigInstallGroups(ctx, branch.ID)
		if err != nil {
			return nil, err
		}

		for j := range groups {
			if groups[j].LabelSelector != nil && groups[j].LabelSelector.Matches(lbls) {
				results = append(results, BranchForLabels{
					Branch: branch,
					Group:  &groups[j],
				})
			}
		}
	}

	return results, nil
}

// ValidateInstallBranchExclusivity checks that an install would not end up
// on two branches after its labels are updated. It returns an error if the
// install is already on a branch and the new labels match a different branch's
// install group.
func (h *Helpers) ValidateInstallBranchExclusivity(ctx context.Context, install *app.Install, newLabels labels.Labels) error {
	matches, err := h.FindBranchesMatchingLabels(ctx, install.AppID, newLabels)
	if err != nil {
		return err
	}

	for _, m := range matches {
		owned, err := h.InstallOwnedByOtherBranch(ctx, install, m.Branch.ID)
		if err != nil {
			return err
		}
		if owned {
			return stderr.ErrUser{
				Err: fmt.Errorf(
					"install %s is on branch %q but labels also match branch %q",
					install.ID, install.AppBranchID.String, m.Branch.Name,
				),
				Description: fmt.Sprintf(
					"This install is already on branch %q. An install can only belong to one app branch at a time. "+
						"Remove it from branch %q first, or change the label selector.",
					install.AppBranchID.String, install.AppBranchID.String,
				),
			}
		}
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Branch.Name
		}
		return stderr.ErrUser{
			Err:         fmt.Errorf("labels match %d branches: %v", len(matches), names),
			Description: "An install can only belong to one app branch at a time. These labels match multiple branches.",
		}
	}

	return nil
}

// ValidateBranchConfigLabelUniqueness checks that no other branch (besides
// excludeBranchID) already uses the same label selector. This prevents two
// deployment plans from matching the same installs.
func (h *Helpers) ValidateBranchConfigLabelUniqueness(ctx context.Context, appID, excludeBranchID string, selectors []*labels.Selector) error {
	var branches []app.AppBranch
	if err := h.db.WithContext(ctx).
		Where(app.AppBranch{AppID: appID}).
		Where("id != ?", excludeBranchID).
		Find(&branches).Error; err != nil {
		return fmt.Errorf("unable to get app branches: %w", err)
	}

	for _, branch := range branches {
		var latestConfig app.AppBranchConfig
		if err := h.db.WithContext(ctx).
			Where(app.AppBranchConfig{AppBranchID: branch.ID}).
			Order("created_at DESC").
			First(&latestConfig).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return fmt.Errorf("unable to get latest config for branch %s: %w", branch.ID, err)
		}

		var groups []app.AppBranchInstallGroup
		if err := h.db.WithContext(ctx).
			Where(app.AppBranchInstallGroup{AppBranchConfigID: latestConfig.ID}).
			Find(&groups).Error; err != nil {
			return fmt.Errorf("unable to get install groups for config %s: %w", latestConfig.ID, err)
		}

		for _, group := range groups {
			if group.LabelSelector == nil {
				continue
			}

			for _, sel := range selectors {
				if sel == nil {
					continue
				}
				if sel.Canonical() == group.LabelSelector.Canonical() {
					return stderr.ErrUser{
						Err: fmt.Errorf(
							"label selector %s already used by branch %q (group %q)",
							sel.Canonical(), branch.Name, group.Name,
						),
						Description: fmt.Sprintf(
							"Branch %q already has an install group with the same label selector. "+
								"Two branches cannot use the same label selector.",
							branch.Name,
						),
					}
				}
			}
		}
	}

	return nil
}

// SyncInstallBranchConnection ensures the install has an active connection
// to the given branch, deactivating any other active connections.
func (h *Helpers) SyncInstallBranchConnection(ctx context.Context, install *app.Install, branchID string) {
	now := time.Now()

	h.db.WithContext(ctx).
		Model(&app.InstallAppBranchConnection{}).
		Where(app.InstallAppBranchConnection{
			InstallID: install.ID,
			Active:    true,
		}).
		Where("app_branch_id != ?", branchID).
		Updates(map[string]interface{}{
			"active":         false,
			"deactivated_at": now,
		})

	var existing app.InstallAppBranchConnection
	err := h.db.WithContext(ctx).
		Where(app.InstallAppBranchConnection{
			InstallID:   install.ID,
			AppBranchID: branchID,
			Active:      true,
		}).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		h.db.WithContext(ctx).Create(&app.InstallAppBranchConnection{
			InstallID:   install.ID,
			AppBranchID: branchID,
			Active:      true,
			ActivatedAt: now,
		})
	}

	h.db.WithContext(ctx).
		Model(install).
		Update("app_branch_id", branchID)
}

// DeactivateInstallBranchConnections deactivates all active branch connections
// for an install and clears its app_branch_id.
func (h *Helpers) DeactivateInstallBranchConnections(ctx context.Context, installID string) {
	now := time.Now()

	h.db.WithContext(ctx).
		Model(&app.InstallAppBranchConnection{}).
		Where(app.InstallAppBranchConnection{
			InstallID: installID,
			Active:    true,
		}).
		Updates(map[string]interface{}{
			"active":         false,
			"deactivated_at": now,
		})

	h.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("id = ?", installID).
		Update("app_branch_id", nil)
}

func (h *Helpers) ReconcileRemovedBranchInstalls(ctx context.Context, appBranchID string, keepInstallIDs []string, selectors []*labels.Selector) error {
	var connected []app.Install
	if err := h.db.WithContext(ctx).
		Where("app_branch_id = ?", appBranchID).
		Find(&connected).Error; err != nil {
		return fmt.Errorf("unable to get branch installs: %w", err)
	}

	keep := make(map[string]struct{}, len(keepInstallIDs))
	for _, id := range keepInstallIDs {
		keep[id] = struct{}{}
	}

	for i := range connected {
		install := &connected[i]
		if _, ok := keep[install.ID]; ok {
			continue
		}

		matched := false
		for _, sel := range selectors {
			if sel != nil && sel.Matches(install.Labels) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		h.DeactivateInstallBranchConnections(ctx, install.ID)
	}

	return nil
}

// ValidateInstallIDsNotOnOtherBranch checks that installs referenced by
// explicit IDs are not already assigned to a different branch.
func (h *Helpers) ValidateInstallIDsNotOnOtherBranch(ctx context.Context, branchID string, installIDs []string) error {
	if len(installIDs) == 0 {
		return nil
	}

	var installs []app.Install
	if err := h.db.WithContext(ctx).
		Where("id IN ?", installIDs).
		Find(&installs).Error; err != nil {
		return fmt.Errorf("unable to get installs: %w", err)
	}

	cache := newOwnerGroupCache()
	for i := range installs {
		install := &installs[i]
		owned, err := h.installOwnedByOtherBranch(ctx, install, branchID, cache)
		if err != nil {
			return err
		}
		if owned {
			return stderr.ErrUser{
				Err: fmt.Errorf(
					"install %q (%s) is already on branch %s",
					install.Name, install.ID, install.AppBranchID.String,
				),
				Description: fmt.Sprintf(
					"Install %q is already assigned to another branch. "+
						"An install can only belong to one branch at a time.",
					install.Name,
				),
			}
		}
	}

	return nil
}

// ClaimSelectorInstallsFromWeakOwners re-points installs a new config's
// selectors match away from a branch that only held them through an
// all-installs group. Without it the handover the validator now allows would
// leave the install still recorded on the branch that let it go.
func (h *Helpers) ClaimSelectorInstallsFromWeakOwners(ctx context.Context, appID, branchID string, selectors []*labels.Selector) error {
	if len(selectors) == 0 {
		return nil
	}

	var installs []app.Install
	if err := h.db.WithContext(ctx).
		Where(app.Install{AppID: appID}).
		Where("app_branch_id IS NOT NULL AND app_branch_id != ?", branchID).
		Find(&installs).Error; err != nil {
		return fmt.Errorf("unable to get installs: %w", err)
	}

	cache := newOwnerGroupCache()
	for i := range installs {
		install := &installs[i]
		if !matchesAny(selectors, install.Labels) {
			continue
		}

		owned, err := h.installOwnedByOtherBranch(ctx, install, branchID, cache)
		if err != nil {
			return err
		}
		if owned {
			continue
		}

		h.SyncInstallBranchConnection(ctx, install, branchID)
	}

	return nil
}

// ValidateBranchConfigInstallsNotOnOtherBranch resolves installs from
// label selectors and checks none are already assigned to a different branch.
func (h *Helpers) ValidateBranchConfigInstallsNotOnOtherBranch(ctx context.Context, appID, branchID string, selectors []*labels.Selector) error {
	cache := newOwnerGroupCache()
	for _, sel := range selectors {
		if sel == nil {
			continue
		}

		var installs []app.Install
		if err := h.db.WithContext(ctx).
			Where(app.Install{AppID: appID}).
			Find(&installs).Error; err != nil {
			return fmt.Errorf("unable to get installs: %w", err)
		}

		for i := range installs {
			install := &installs[i]
			if !sel.Matches(install.Labels) {
				continue
			}
			owned, err := h.installOwnedByOtherBranch(ctx, install, branchID, cache)
			if err != nil {
				return err
			}
			if owned {
				return stderr.ErrUser{
					Err: fmt.Errorf(
						"install %q (%s) is already on branch %s",
						install.Name, install.ID, install.AppBranchID.String,
					),
					Description: fmt.Sprintf(
						"Install %q is already assigned to another branch. "+
							"An install can only belong to one branch at a time.",
						install.Name,
					),
				}
			}
		}
	}

	return nil
}
