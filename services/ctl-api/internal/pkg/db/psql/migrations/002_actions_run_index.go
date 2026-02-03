package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

func (m *Migrations) Migration002DropOldActionsRunIndex(ctx context.Context, db *gorm.DB) error {
	dropSQL := `DROP INDEX IF EXISTS idx_iawr_iaw_id_delete_id_created_at`

	if res := db.WithContext(ctx).
		Exec(dropSQL); res.Error != nil {
		return res.Error
	}

	return nil
}
