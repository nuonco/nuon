package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

func (m *Migrations) Migration141BackfillInstallAppBranchGroupAssignmentSource(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).
		Model(&app.InstallAppBranchConnection{}).
		Where("COALESCE(app_branch_group, '') != ''").
		Where("COALESCE(app_branch_group_assignment_source, '') = ''").
		Update("app_branch_group_assignment_source", app.InstallAppBranchGroupAssignmentSourceExplicit).Error; err != nil {
		return fmt.Errorf("unable to backfill explicit app branch group assignments: %w", err)
	}

	var connections []app.InstallAppBranchConnection
	if err := db.WithContext(ctx).
		Where(app.InstallAppBranchConnection{
			Active: true,
		}).
		Where("COALESCE(app_branch_group, '') = ''").
		Where("COALESCE(app_branch_group_assignment_source, '') = ''").
		Find(&connections).Error; err != nil {
		return fmt.Errorf("unable to load derived app branch group assignments: %w", err)
	}

	for idx := range connections {
		connection := &connections[idx]
		var install app.Install
		if err := db.WithContext(ctx).First(&install, "id = ?", connection.InstallID).Error; err != nil {
			return fmt.Errorf("unable to load install %s: %w", connection.InstallID, err)
		}
		install.AppBranchGroup = ""
		install.AppBranchGroupAssignmentSource = ""

		groups, err := appshelpers.LatestConfigInstallGroupsWithDB(ctx, db, connection.AppBranchID)
		if err != nil {
			return err
		}
		group, source, err := appshelpers.ResolveInstallGroupAssignment(groups, &install)
		if err != nil {
			return err
		}
		if group == nil {
			continue
		}
		if err := db.WithContext(ctx).
			Model(connection).
			Updates(map[string]any{
				"app_branch_group":                   group.Name,
				"app_branch_group_assignment_source": source,
			}).Error; err != nil {
			return fmt.Errorf("unable to backfill install %s app branch group assignment: %w", connection.InstallID, err)
		}
	}

	return nil
}
