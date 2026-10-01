package migrations

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migration142MergeNewAppIAFeatures folds new-install-ia into new-app-ia so orgs
// that had the retired install IA flag keep that dashboard behavior.
func (m *Migrations) Migration142MergeNewAppIAFeatures(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`
		UPDATE orgs
		SET features = jsonb_set(COALESCE(features, '{}'::jsonb), '{new-app-ia}', 'true'::jsonb, true),
		    updated_at = now()
		WHERE COALESCE(features->>'new-app-ia', 'false') <> 'true'
		  AND features->>'new-install-ia' = 'true';`)
	if res.Error != nil {
		return res.Error
	}

	m.l.Info("merged new app ia feature flags", zap.Int64("rows", res.RowsAffected))
	return nil
}
