package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration109BackfillNuonAccountsIsAdmin grants admin access (is_admin = true) to all existing
// Nuon employee accounts. It mirrors the legacy IsEmployee definition — auth/auth0 accounts whose
// email ends in @nuon.co — so the new persisted flag matches the prior email-based admin check.
// Idempotent: only flips accounts that aren't already admins.
func (m *Migrations) Migration109BackfillNuonAccountsIsAdmin(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).Exec(`
		UPDATE accounts
		SET is_admin = true
		WHERE account_type IN ('auth', 'auth0')
		  AND email LIKE '%@nuon.co'
		  AND is_admin = false
		  AND deleted_at = 0;
	`); res.Error != nil {
		return res.Error
	}

	return nil
}
