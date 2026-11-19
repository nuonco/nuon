package migrations

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

func (m *Migrations) Migration159DropInstallSandboxConfigID(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`DROP VIEW IF EXISTS installs_migration_test;
DROP VIEW IF EXISTS installs_view_v9;
ALTER TABLE installs DROP COLUMN IF EXISTS app_sandbox_config_id;
CREATE VIEW installs_view_v9 AS
` + viewsql.InstallsViewV9).Error
}
