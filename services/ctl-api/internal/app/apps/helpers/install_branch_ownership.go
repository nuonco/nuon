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

func (h *Helpers) SetInstallAppBranch(ctx context.Context, installID, branchID string) error {
	return h.SetInstallAppBranchGroup(ctx, installID, branchID, "")
}

func (h *Helpers) SetInstallAppBranchGroup(ctx context.Context, installID, branchID, group string) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return SetInstallAppBranchGroupWithDB(ctx, tx, installID, branchID, group)
	})
}

func SetInstallAppBranchWithDB(ctx context.Context, db *gorm.DB, installID, branchID string) error {
	return SetInstallAppBranchGroupWithDB(ctx, db, installID, branchID, "")
}

func SetInstallAppBranchGroupWithDB(ctx context.Context, db *gorm.DB, installID, branchID, group string) error {
	now := time.Now()

	if err := db.WithContext(ctx).
		Model(&app.InstallAppBranchConnection{}).
		Where(app.InstallAppBranchConnection{InstallID: installID, Active: true}).
		Where("app_branch_id != ? OR COALESCE(app_branch_group, '') != ?", branchID, group).
		Updates(map[string]any{
			"active":         false,
			"deactivated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("unable to deactivate install branch connections: %w", err)
	}

	var existing app.InstallAppBranchConnection
	err := db.WithContext(ctx).
		Where("install_id = ? AND app_branch_id = ? AND COALESCE(app_branch_group, '') = ? AND active = ?", installID, branchID, group, true).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if err := db.WithContext(ctx).Create(&app.InstallAppBranchConnection{
			InstallID:      installID,
			AppBranchID:    branchID,
			AppBranchGroup: group,
			Active:         true,
			ActivatedAt:    now,
		}).Error; err != nil {
			return fmt.Errorf("unable to create install branch connection: %w", err)
		}
	case err != nil:
		return fmt.Errorf("unable to load install branch connection: %w", err)
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
		Joins("JOIN install_app_branch_connections ON install_app_branch_connections.install_id = installs.id AND install_app_branch_connections.active = ? AND install_app_branch_connections.deleted_at = 0", true).
		Where("install_app_branch_connections.app_branch_id = ?", branchID).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to load installs for branch %s: %w", branchID, err)
	}
	return installs, nil
}

func InstallMatchesGroup(group *app.AppBranchInstallGroup, install *app.Install) bool {
	if group == nil {
		return false
	}
	if install.AppBranchGroup != "" {
		return group.Name == install.AppBranchGroup
	}
	if group.LabelSelector != nil && len(group.LabelSelector.MatchLabels) > 0 {
		return group.LabelSelector.Matches(install.Labels)
	}
	return false
}

func InstallGroupsMatching(groups []app.AppBranchInstallGroup, install *app.Install) []string {
	var names []string
	for i := range groups {
		if InstallMatchesGroup(&groups[i], install) {
			names = append(names, groups[i].Name)
		}
	}
	return names
}

func ResolveInstallGroup(groups []app.AppBranchInstallGroup, install *app.Install) (*app.AppBranchInstallGroup, error) {
	if install.AppBranchGroup != "" {
		for i := range groups {
			if groups[i].Name == install.AppBranchGroup {
				return &groups[i], nil
			}
		}
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("install %s selects unknown app branch group %s", install.ID, install.AppBranchGroup),
			Description: fmt.Sprintf("The selected app branch group %q does not exist.", install.AppBranchGroup),
		}
	}

	matched := InstallGroupsMatching(groups, install)
	if len(matched) > 1 {
		sort.Strings(matched)
		name := install.Name
		if name == "" {
			name = install.ID
		}
		return nil, stderr.ErrUser{
			Err: fmt.Errorf("install %s matches install groups %s", install.ID, strings.Join(matched, ", ")),
			Description: fmt.Sprintf(
				"Install %q matches more than one install group (%s). An install can only belong to one install group on a branch.",
				name, strings.Join(matched, ", "),
			),
		}
	}
	if len(matched) == 1 {
		for i := range groups {
			if groups[i].Name == matched[0] {
				return &groups[i], nil
			}
		}
	}
	for i := range groups {
		if groups[i].Default {
			return &groups[i], nil
		}
	}
	return nil, nil
}

func ValidateInstallSingleGroup(groups []app.AppBranchInstallGroup, install *app.Install) error {
	_, err := ResolveInstallGroup(groups, install)
	return err
}

func (h *Helpers) ValidateBranchInstallsSingleGroup(ctx context.Context, branchID string, groups []app.AppBranchInstallGroup) error {
	return ValidateBranchInstallsSingleGroupWithDB(ctx, h.db, branchID, groups)
}

// Callers inside a transaction must use this so they see their own writes.
func ValidateBranchInstallsSingleGroupWithDB(ctx context.Context, db *gorm.DB, branchID string, groups []app.AppBranchInstallGroup) error {
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
