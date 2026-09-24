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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func (h *Helpers) SetInstallAppBranch(ctx context.Context, installID, branchID string) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var install app.Install
		if err := tx.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
			return fmt.Errorf("unable to load install: %w", err)
		}
		groups, err := LatestConfigInstallGroupsWithDB(ctx, tx, branchID)
		if err != nil {
			return err
		}
		install.AppBranchGroup = ""
		install.AppBranchGroupAssignmentSource = ""
		group, source, err := ResolveInstallGroupAssignment(groups, &install)
		if err != nil {
			return err
		}
		if group == nil {
			return NoMatchingInstallGroupError(&install)
		}
		return SetInstallAppBranchGroupAssignmentWithDB(ctx, tx, installID, branchID, group.Name, source)
	})
}

func (h *Helpers) SetInstallAppBranchGroup(ctx context.Context, installID, branchID, group string) error {
	source := app.InstallAppBranchGroupAssignmentSource("")
	if group != "" {
		source = app.InstallAppBranchGroupAssignmentSourceExplicit
	}
	return h.SetInstallAppBranchGroupAssignment(ctx, installID, branchID, group, source)
}

func (h *Helpers) SetInstallAppBranchGroupAssignment(ctx context.Context, installID, branchID, group string, source app.InstallAppBranchGroupAssignmentSource) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return SetInstallAppBranchGroupAssignmentWithDB(ctx, tx, installID, branchID, group, source)
	})
}

func SetInstallAppBranchWithDB(ctx context.Context, db *gorm.DB, installID, branchID string) error {
	return SetInstallAppBranchGroupWithDB(ctx, db, installID, branchID, "")
}

func SetInstallAppBranchGroupWithDB(ctx context.Context, db *gorm.DB, installID, branchID, group string) error {
	source := app.InstallAppBranchGroupAssignmentSource("")
	if group != "" {
		source = app.InstallAppBranchGroupAssignmentSourceExplicit
	}
	return SetInstallAppBranchGroupAssignmentWithDB(ctx, db, installID, branchID, group, source)
}

func SetInstallAppBranchGroupAssignmentWithDB(ctx context.Context, db *gorm.DB, installID, branchID, group string, source app.InstallAppBranchGroupAssignmentSource) error {
	now := time.Now()

	if err := db.WithContext(ctx).
		Model(&app.InstallAppBranchConnection{}).
		Where(app.InstallAppBranchConnection{InstallID: installID, Active: true}).
		Where("app_branch_id != ? OR COALESCE(app_branch_group, '') != ? OR COALESCE(app_branch_group_assignment_source, '') != ?", branchID, group, source).
		Updates(map[string]any{
			"active":         false,
			"deactivated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("unable to deactivate install branch connections: %w", err)
	}

	var existing app.InstallAppBranchConnection
	err := db.WithContext(ctx).
		Where(app.InstallAppBranchConnection{
			InstallID:                      installID,
			AppBranchID:                    branchID,
			AppBranchGroup:                 group,
			AppBranchGroupAssignmentSource: source,
			Active:                         true,
		}).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if err := db.WithContext(ctx).Create(&app.InstallAppBranchConnection{
			InstallID:                      installID,
			AppBranchID:                    branchID,
			AppBranchGroup:                 group,
			AppBranchGroupAssignmentSource: source,
			Active:                         true,
			ActivatedAt:                    now,
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
	installIDCol := views.TableOrViewName(db, &app.Install{}, ".id")
	if err := db.WithContext(ctx).
		Joins("JOIN install_app_branch_connections ON install_app_branch_connections.install_id = "+installIDCol+" AND install_app_branch_connections.active = ? AND install_app_branch_connections.deleted_at = 0", true).
		Where("install_app_branch_connections.app_branch_id = ?", branchID).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to load installs for branch %s: %w", branchID, err)
	}
	return installs, nil
}

func installUsesExplicitGroup(install *app.Install) bool {
	return install.AppBranchGroupAssignmentSource == app.InstallAppBranchGroupAssignmentSourceExplicit ||
		(install.AppBranchGroupAssignmentSource == "" && install.AppBranchGroup != "")
}

func NoMatchingInstallGroupError(install *app.Install) error {
	name := install.Name
	if name == "" {
		name = install.ID
	}
	return stderr.ErrUser{
		Err:         fmt.Errorf("install %s does not match an app branch group", install.ID),
		Description: fmt.Sprintf("Install %q does not match an install group and the branch has no default group.", name),
	}
}

func InstallMatchesGroup(group *app.AppBranchInstallGroup, install *app.Install) bool {
	if group == nil {
		return false
	}
	if installUsesExplicitGroup(install) {
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
	group, _, err := ResolveInstallGroupAssignment(groups, install)
	return group, err
}

func ResolveInstallGroupAssignment(groups []app.AppBranchInstallGroup, install *app.Install) (*app.AppBranchInstallGroup, app.InstallAppBranchGroupAssignmentSource, error) {
	if installUsesExplicitGroup(install) {
		for i := range groups {
			if groups[i].Name == install.AppBranchGroup {
				return &groups[i], app.InstallAppBranchGroupAssignmentSourceExplicit, nil
			}
		}
		return nil, "", stderr.ErrUser{
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
		return nil, "", stderr.ErrUser{
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
				return &groups[i], app.InstallAppBranchGroupAssignmentSourceLabels, nil
			}
		}
	}
	for i := range groups {
		if groups[i].Default {
			return &groups[i], app.InstallAppBranchGroupAssignmentSourceDefault, nil
		}
	}
	return nil, "", nil
}

func ReconcileInstallAppBranchGroupWithDB(ctx context.Context, db *gorm.DB, installID string) error {
	var install app.Install
	if err := db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		return fmt.Errorf("unable to load install: %w", err)
	}
	if !install.AppBranchID.Valid || install.AppBranchID.String == "" || installUsesExplicitGroup(&install) {
		return nil
	}

	groups, err := LatestConfigInstallGroupsWithDB(ctx, db, install.AppBranchID.String)
	if err != nil {
		return err
	}
	group, source, err := ResolveInstallGroupAssignment(groups, &install)
	if err != nil {
		return err
	}
	if group == nil {
		return NoMatchingInstallGroupError(&install)
	}
	return SetInstallAppBranchGroupAssignmentWithDB(ctx, db, install.ID, install.AppBranchID.String, group.Name, source)
}

func (h *Helpers) ReconcileInstallAppBranchGroup(ctx context.Context, installID string) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return ReconcileInstallAppBranchGroupWithDB(ctx, tx, installID)
	})
}

func ReconcileBranchInstallAppBranchGroupsWithDB(ctx context.Context, db *gorm.DB, branchID string, groups []app.AppBranchInstallGroup) error {
	installs, err := BranchInstallsWithDB(ctx, db, branchID)
	if err != nil {
		return err
	}
	for idx := range installs {
		install := &installs[idx]
		if installUsesExplicitGroup(install) {
			continue
		}
		group, source, err := ResolveInstallGroupAssignment(groups, install)
		if err != nil {
			return err
		}
		if group == nil {
			continue
		}
		if err := SetInstallAppBranchGroupAssignmentWithDB(ctx, db, install.ID, branchID, group.Name, source); err != nil {
			return err
		}
	}
	return nil
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
		if installUsesExplicitGroup(&installs[i]) {
			continue
		}
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
