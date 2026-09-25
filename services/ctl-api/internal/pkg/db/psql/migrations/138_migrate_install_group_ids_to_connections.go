package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (m *Migrations) Migration138MigrateInstallGroupIDsToConnections(ctx context.Context, db *gorm.DB) error {
	if !db.Migrator().HasColumn("app_branch_install_groups", "install_ids") {
		return nil
	}

	if err := db.WithContext(ctx).Exec(`
UPDATE install_app_branch_connections AS connection
SET app_branch_group = install_group.name
FROM app_branch_install_groups AS install_group
JOIN app_branch_configs AS branch_config
  ON branch_config.id = install_group.app_branch_config_id
WHERE connection.active = TRUE
  AND connection.app_branch_id = branch_config.app_branch_id
  AND connection.install_id = ANY(install_group.install_ids)
  AND branch_config.id = (
    SELECT latest.id
    FROM app_branch_configs AS latest
    WHERE latest.app_branch_id = connection.app_branch_id
      AND latest.deleted_at = 0
    ORDER BY latest.created_at DESC, latest.id DESC
    LIMIT 1
  )`).Error; err != nil {
		return fmt.Errorf("unable to migrate explicit install groups to app branch connections: %w", err)
	}
	if err := db.WithContext(ctx).Exec("ALTER TABLE app_branch_install_groups DROP COLUMN install_ids").Error; err != nil {
		return fmt.Errorf("unable to drop app branch install group install_ids: %w", err)
	}

	return nil
}
