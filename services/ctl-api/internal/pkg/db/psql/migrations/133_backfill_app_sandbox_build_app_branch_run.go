package migrations

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (m *Migrations) Migration133BackfillAppSandboxBuildAppBranchRun(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`
		UPDATE app_sandbox_builds AS b
		SET app_branch_run_id = (
			SELECT r.id
			FROM app_branch_runs r
			WHERE r.app_config_id = b.app_config_id
			  AND r.deleted_at = 0
			ORDER BY (r.created_at <= b.created_at) DESC,
			         CASE
			             WHEN r.created_at <= b.created_at THEN b.created_at - r.created_at
			             ELSE r.created_at - b.created_at
			         END ASC,
			         r.id ASC
			LIMIT 1
		),
		    updated_at = now()
		WHERE b.app_branch_run_id IS NULL
		  AND b.app_config_id IS NOT NULL
		  AND b.app_config_id <> ''
		  AND EXISTS (
			SELECT 1
			FROM app_branch_runs r
			WHERE r.app_config_id = b.app_config_id
			  AND r.deleted_at = 0
		  );`)
	if res.Error != nil {
		return res.Error
	}

	m.l.Info("backfilled app sandbox build branch runs", zap.Int64("rows", res.RowsAffected))
	return nil
}
