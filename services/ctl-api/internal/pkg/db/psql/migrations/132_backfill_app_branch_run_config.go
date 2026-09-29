package migrations

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (m *Migrations) Migration132BackfillAppBranchRunConfig(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`
		UPDATE app_branch_configs
		SET run_config = '{"mode":"push"}'::jsonb,
		    updated_at = now()
		WHERE run_config IS NULL
		   OR run_config = 'null'::jsonb;`)
	if res.Error != nil {
		return res.Error
	}

	m.l.Info("backfilled app branch run configs", zap.Int64("rows", res.RowsAffected))
	return nil
}
