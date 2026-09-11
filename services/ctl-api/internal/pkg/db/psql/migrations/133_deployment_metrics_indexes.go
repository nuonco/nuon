package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration133DeploymentMetricsIndexes(ctx context.Context, db *gorm.DB) error {
	for _, query := range []string{
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_deploys_metrics_created
ON install_deploys (created_at)
WHERE deleted_at = 0 AND type = 'apply'`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_deploys_metrics_applied
ON install_deploys (applied_at)
WHERE deleted_at = 0 AND type = 'apply' AND applied_at IS NOT NULL`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_deploys_metrics_latest
ON install_deploys (install_component_id, created_at DESC, id DESC)
WHERE deleted_at = 0 AND type = 'apply'`,
	} {
		if err := db.WithContext(ctx).Exec(query).Error; err != nil {
			return err
		}
	}
	return nil
}
