package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (m *Migrations) Migration139MigrateAllInstallsToDefault(ctx context.Context, db *gorm.DB) error {
	if !db.Migrator().HasColumn("app_branch_install_groups", "all_installs") {
		return nil
	}

	if err := db.WithContext(ctx).
		Exec("UPDATE app_branch_install_groups SET \"default\" = TRUE WHERE all_installs = TRUE").Error; err != nil {
		return fmt.Errorf("unable to migrate default install groups: %w", err)
	}
	if err := db.WithContext(ctx).Exec("ALTER TABLE app_branch_install_groups DROP COLUMN all_installs").Error; err != nil {
		return fmt.Errorf("unable to drop app branch install group all_installs: %w", err)
	}

	return nil
}
