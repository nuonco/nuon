package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration100FixApprovalOptionDefault(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(`UPDATE install_configs SET approval_option = 'prompt' WHERE approval_option = 'auto' AND deleted_at = 0;`); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(`UPDATE install_workflows SET approval_option = 'prompt' WHERE approval_option = 'auto' AND deleted_at = 0;`); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(`ALTER TABLE install_configs ALTER COLUMN approval_option SET DEFAULT 'prompt';`); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(`ALTER TABLE install_workflows ALTER COLUMN approval_option SET DEFAULT 'prompt';`); res.Error != nil {
		return res.Error
	}

	return nil
}
