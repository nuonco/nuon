package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration105FixSkipNoopsDefault(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`UPDATE component_config_connections SET skip_noops = false WHERE skip_noops = true;`)
	return res.Error
}
