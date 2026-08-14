package migrations

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migration127BackfillOrgEmailNotifications enables email notifications on org
// notifications configs that were created between AccountTypeAuth landing
// (2026-01-10, #128) and create_org widening its check to accept it
// (2026-06-03, #1573). Orgs created by an `auth` account in that window got
// enable_email_notifications=false, which silently drops every org invite email
// because notifications.SendEmail returns nil when the flag is off.
func (m *Migrations) Migration127BackfillOrgEmailNotifications(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`
		UPDATE notifications_configs AS nc
		SET enable_email_notifications = true,
		    updated_at = now()
		FROM orgs AS o
		JOIN accounts AS a ON a.id = o.created_by_id
		WHERE nc.owner_id = o.id
		  AND nc.owner_type = 'orgs'
		  AND nc.deleted_at = 0
		  AND nc.enable_email_notifications = false
		  AND a.account_type IN ('auth', 'auth0');`)
	if res.Error != nil {
		return res.Error
	}

	m.l.Info("backfilled org email notifications", zap.Int64("rows", res.RowsAffected))
	return nil
}
