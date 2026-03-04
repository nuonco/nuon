package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration094DropCollidingIndexNames(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).Exec(`DROP INDEX IF EXISTS idx_owner;`); res.Error != nil {
		return res.Error
	}

	// The original struct tag used `index:idx_account_role:unique` (colon instead of comma),
	// so GORM treated `:unique` as part of the literal name and created a non-unique index.
	// Uniqueness was never enforced and duplicate (role_id, account_id) rows exist.
	// Drop the malformed index here; a follow-up PR will dedup data and add a proper unique index.
	if res := db.WithContext(ctx).Exec(`DROP INDEX IF EXISTS "idx_account_role:unique";`); res.Error != nil {
		return res.Error
	}

	return nil
}
