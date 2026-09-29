package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration102WebhooksMatch(ctx context.Context, db *gorm.DB) error {
	stmts := []string{
		`DROP INDEX IF EXISTS idx_webhooks_org_url;`,
	}
	for _, qry := range stmts {
		if res := db.WithContext(ctx).Exec(qry); res.Error != nil {
			return res.Error
		}
	}
	return nil
}
