package helpers

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// An install can be unbranched or belong to exactly one app branch. Ownership
// is recorded on Install.AppBranchID and mirrored by an active
// InstallAppBranchConnection. It only changes when the install is created
// against a branch or explicitly moved to another one. Labels, install group
// selectors and branch config saves only decide group membership within the
// owning branch.

// SetInstallAppBranch makes branchID the install's owner, deactivating any
// other active connection. Both halves of the record move in one transaction so
// a failure cannot leave the pin and the connection disagreeing.
func (h *Helpers) SetInstallAppBranch(ctx context.Context, installID, branchID string) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return SetInstallAppBranchWithDB(ctx, tx, installID, branchID)
	})
}

// SetInstallAppBranchWithDB updates both ownership records using the caller's
// transaction. It is used during install creation so the install row and its
// ownership connection cannot be committed separately.
func SetInstallAppBranchWithDB(ctx context.Context, db *gorm.DB, installID, branchID string) error {
	now := time.Now()

	if err := db.WithContext(ctx).
		Model(&app.InstallAppBranchConnection{}).
		Where(app.InstallAppBranchConnection{InstallID: installID, Active: true}).
		Where("app_branch_id != ?", branchID).
		Updates(map[string]any{
			"active":         false,
			"deactivated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("unable to deactivate install branch connections: %w", err)
	}

	var existing app.InstallAppBranchConnection
	err := db.WithContext(ctx).
		Where(app.InstallAppBranchConnection{
			InstallID:   installID,
			AppBranchID: branchID,
			Active:      true,
		}).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if err := db.WithContext(ctx).Create(&app.InstallAppBranchConnection{
			InstallID:   installID,
			AppBranchID: branchID,
			Active:      true,
			ActivatedAt: now,
		}).Error; err != nil {
			return fmt.Errorf("unable to create install branch connection: %w", err)
		}
	case err != nil:
		return fmt.Errorf("unable to load install branch connection: %w", err)
	}

	if err := db.WithContext(ctx).
		Model(&app.Install{}).
		Where(app.Install{ID: installID}).
		Update("app_branch_id", branchID).Error; err != nil {
		return fmt.Errorf("unable to pin install to app branch: %w", err)
	}

	return nil
}

// BranchInstalls returns the installs the branch owns, which is the only
// population any of its install groups can resolve to.
func (h *Helpers) BranchInstalls(ctx context.Context, branchID string) ([]app.Install, error) {
	return BranchInstallsWithDB(ctx, h.db, branchID)
}

// Callers inside a transaction must use this so they see their own writes.
func BranchInstallsWithDB(ctx context.Context, db *gorm.DB, branchID string) ([]app.Install, error) {
	var installs []app.Install
	if err := db.WithContext(ctx).
		Where("app_branch_id = ?", branchID).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to load installs for branch %s: %w", branchID, err)
	}
	return installs, nil
}

// AppInstalls returns every install on the app, regardless of which branch
// owns it. Unlike BranchInstalls, this is not the target population for
// install groups; it exists for callers that intentionally reach across
// branches, such as resolving a preview target.
func (h *Helpers) AppInstalls(ctx context.Context, appID string) ([]app.Install, error) {
	var installs []app.Install
	if err := h.db.WithContext(ctx).
		Where(app.Install{AppID: appID}).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to load installs for app %s: %w", appID, err)
	}
	return installs, nil
}

// InstallMatchesGroup reports whether a group targets an install the branch
// already owns. Ownership is the caller's responsibility: a group never reaches
// outside its branch, so passing an install another branch owns is a bug.
func InstallMatchesGroup(group *app.AppBranchInstallGroup, install *app.Install) bool {
	if group == nil {
		return false
	}
	if group.AllInstalls {
		return true
	}
	if group.LabelSelector != nil && len(group.LabelSelector.MatchLabels) > 0 {
		return group.LabelSelector.Matches(install.Labels)
	}
	for _, id := range group.InstallIDs {
		if id == install.ID {
			return true
		}
	}
	return false
}

// InstallGroupsMatching names every group in the set that targets the install.
func InstallGroupsMatching(groups []app.AppBranchInstallGroup, install *app.Install) []string {
	var names []string
	for i := range groups {
		if InstallMatchesGroup(&groups[i], install) {
			names = append(names, groups[i].Name)
		}
	}
	return names
}

// ValidateInstallSingleGroup rejects an install that more than one of a
// branch's groups targets. Two groups holding the same install would plan and
// deploy it twice in the same branch run, in an order nothing defines.
func ValidateInstallSingleGroup(groups []app.AppBranchInstallGroup, install *app.Install) error {
	matched := InstallGroupsMatching(groups, install)
	if len(matched) < 2 {
		return nil
	}

	sort.Strings(matched)
	name := install.Name
	if name == "" {
		name = install.ID
	}

	return stderr.ErrUser{
		Err: fmt.Errorf("install %s matches install groups %s", install.ID, strings.Join(matched, ", ")),
		Description: fmt.Sprintf(
			"Install %q matches more than one install group (%s). An install can only belong to one install group on a branch.",
			name, strings.Join(matched, ", "),
		),
	}
}

// ValidateBranchInstallsSingleGroup checks the installs a branch owns against
// the groups a config is about to save.
func (h *Helpers) ValidateBranchInstallsSingleGroup(ctx context.Context, branchID string, groups []app.AppBranchInstallGroup) error {
	return ValidateBranchInstallsSingleGroupWithDB(ctx, h.db, branchID, groups)
}

// Callers inside a transaction must use this so they see their own writes.
func ValidateBranchInstallsSingleGroupWithDB(ctx context.Context, db *gorm.DB, branchID string, groups []app.AppBranchInstallGroup) error {
	if len(groups) < 2 {
		return nil
	}

	installs, err := BranchInstallsWithDB(ctx, db, branchID)
	if err != nil {
		return err
	}

	for i := range installs {
		if err := ValidateInstallSingleGroup(groups, &installs[i]); err != nil {
			return err
		}
	}

	return nil
}

// ValidateInstallLabelsSingleGroup checks an install's proposed labels against
// the groups on the branch that owns it. Labels only decide group membership
// now, so a label change that lands the install in two groups is rejected
// rather than silently moving it anywhere.
func (h *Helpers) ValidateInstallLabelsSingleGroup(ctx context.Context, install *app.Install, newLabels labels.Labels) error {
	branchID := install.AppBranchID.String
	if !install.AppBranchID.Valid || branchID == "" {
		return nil
	}

	groups, err := h.LatestConfigInstallGroups(ctx, branchID)
	if err != nil {
		return err
	}

	candidate := *install
	candidate.Labels = newLabels
	return ValidateInstallSingleGroup(groups, &candidate)
}

// ValidateInstallIDsOwnedByBranch rejects a config that names installs the
// branch does not own. Listing an install by ID selects it for a group; it no
// longer claims it, so the install has to be created on the branch or moved
// there first.
func (h *Helpers) ValidateInstallIDsOwnedByBranch(ctx context.Context, branchID string, installIDs []string) error {
	return ValidateInstallIDsOwnedByBranchWithDB(ctx, h.db, branchID, installIDs)
}

// Callers inside a transaction must use this so they see their own writes.
func ValidateInstallIDsOwnedByBranchWithDB(ctx context.Context, db *gorm.DB, branchID string, installIDs []string) error {
	if len(installIDs) == 0 {
		return nil
	}

	var installs []app.Install
	if err := db.WithContext(ctx).
		Where("id IN ?", installIDs).
		Find(&installs).Error; err != nil {
		return fmt.Errorf("unable to get installs: %w", err)
	}

	byID := make(map[string]*app.Install, len(installs))
	for i := range installs {
		byID[installs[i].ID] = &installs[i]
	}

	for _, id := range installIDs {
		install, found := byID[id]
		if !found {
			return stderr.ErrUser{
				Err:         fmt.Errorf("install %q not found", id),
				Description: fmt.Sprintf("Install %q does not exist.", id),
			}
		}
		if install.AppBranchID.Valid && install.AppBranchID.String == branchID {
			continue
		}

		if !install.AppBranchID.Valid || install.AppBranchID.String == "" {
			return stderr.ErrUser{
				Err:         fmt.Errorf("install %q (%s) is not on branch %s", install.Name, install.ID, branchID),
				Description: fmt.Sprintf("Install %q is not on this branch. Move it to this branch before adding it to an install group.", install.Name),
			}
		}

		return stderr.ErrUser{
			Err: fmt.Errorf("install %q (%s) is on branch %s", install.Name, install.ID, install.AppBranchID.String),
			Description: fmt.Sprintf(
				"Install %q belongs to another app branch. Move it to this branch before adding it to an install group.",
				install.Name,
			),
		}
	}

	return nil
}
