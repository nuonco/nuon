package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration105FixSkipNoopsDefault resets skip_noops to false for all component
// config connections that were created with the incorrect default of true.
func (m *Migrations) Migration105FixSkipNoopsDefault(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`UPDATE component_config_connections SET skip_noops = false WHERE skip_noops = true;`)
	return res.Error
}
