package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration100FixApprovalOptionDefault fixes the approval_option column default
// from 'auto' (invalid) to 'prompt' and backfills existing rows.
func (m *Migrations) Migration100FixApprovalOptionDefault(ctx context.Context, db *gorm.DB) error {
	// Fix existing install_configs with invalid 'auto' value
	if res := db.WithContext(ctx).
		Exec(`UPDATE install_configs SET approval_option = 'prompt' WHERE approval_option = 'auto' AND deleted_at = 0;`); res.Error != nil {
		return res.Error
	}

	// Fix existing workflows with invalid 'auto' value
	if res := db.WithContext(ctx).
		Exec(`UPDATE install_workflows SET approval_option = 'prompt' WHERE approval_option = 'auto' AND deleted_at = 0;`); res.Error != nil {
		return res.Error
	}

	// Update column defaults
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
