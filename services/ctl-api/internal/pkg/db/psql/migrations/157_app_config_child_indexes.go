package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration157AppConfigChildIndexes(ctx context.Context, db *gorm.DB) error {
	for name, columns := range map[string]string{
		"idx_app_permissions_configs_app_config_id":    "app_permissions_configs (app_config_id, deleted_at)",
		"idx_app_break_glass_configs_app_config_id":    "app_break_glass_configs (app_config_id, deleted_at)",
		"idx_app_secrets_configs_app_config_id":        "app_secrets_configs (app_config_id, deleted_at)",
		"idx_app_stack_configs_app_config_id":          "app_stack_configs (app_config_id, deleted_at)",
		"idx_app_runner_configs_app_config_id":         "app_runner_configs (app_config_id, deleted_at)",
		"idx_app_awsiam_role_configs_owner_id":         "app_awsiam_role_configs (owner_id, deleted_at)",
		"idx_app_secret_configs_app_secrets_config_id": "app_secret_configs (app_secrets_config_id, deleted_at)",
	} {
		if err := ensureValidUniqueIndex(ctx, db, name, "CREATE INDEX CONCURRENTLY IF NOT EXISTS "+name+" ON "+columns); err != nil {
			return err
		}
	}
	return nil
}
