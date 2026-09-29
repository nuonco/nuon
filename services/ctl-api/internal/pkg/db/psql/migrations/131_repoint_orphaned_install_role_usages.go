package migrations

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

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
