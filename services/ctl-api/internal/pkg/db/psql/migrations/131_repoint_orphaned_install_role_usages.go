package migrations

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migration131RepointOrphanedInstallRoleUsages moves usage rows off install_roles
// that MigrateInstallRoles used to soft-delete and recreate on every app config
// sync, onto the live row for the same install and role name. The last-used
// lookup only reads live rows, so without this it stays empty for every role
// used before the recreate stopped.
func (m *Migrations) Migration131RepointOrphanedInstallRoleUsages(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`
		UPDATE install_role_usages AS u
		SET install_role_id = live.id,
		    updated_at = now()
		FROM install_roles AS dead
		JOIN app_awsiam_role_configs AS dc ON dc.id = dead.app_role_config_id
		JOIN install_roles AS live
		  ON live.install_id = dead.install_id
		 AND live.deleted_at = 0
		JOIN app_awsiam_role_configs AS lc
		  ON lc.id = live.app_role_config_id
		 AND lc.name = dc.name
		WHERE u.install_role_id = dead.id
		  AND dead.deleted_at <> 0;`)
	if res.Error != nil {
		return res.Error
	}

	m.l.Info("repointed orphaned install role usages", zap.Int64("rows", res.RowsAffected))
	return nil
}
