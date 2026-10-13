package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 143_install_state_invalidation_triggers.sql
var InstallStateInvalidationTriggersSQL string

func (m *Migrations) Migration143InstallStateInvalidationTriggers(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(InstallStateInvalidationTriggersSQL); res.Error != nil {
		return res.Error
	}

	return nil
}
