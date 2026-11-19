package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration158AppBranchRunConfigPathsCheck makes run_config.paths (allow list)
// and run_config.paths_ignore (deny list) mutually exclusive, and rejects
// negated "!" patterns in either list.
func (m *Migrations) Migration158AppBranchRunConfigPathsCheck(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'chk_app_branch_configs_run_config_paths'
				  AND conrelid = 'app_branch_configs'::regclass
			) THEN
				ALTER TABLE app_branch_configs
					ADD CONSTRAINT chk_app_branch_configs_run_config_paths CHECK (
						NOT (
							jsonb_path_exists(run_config, '$.paths[0]')
							AND jsonb_path_exists(run_config, '$.paths_ignore[0]')
						)
						AND NOT jsonb_path_exists(run_config, '$.paths[*] ? (@ starts with "!")')
						AND NOT jsonb_path_exists(run_config, '$.paths_ignore[*] ? (@ starts with "!")')
					) NOT VALID;
			END IF;
		END
		$$;
		ALTER TABLE app_branch_configs VALIDATE CONSTRAINT chk_app_branch_configs_run_config_paths;`).Error
}
