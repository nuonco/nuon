package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration152DedupeAccountRoles(ctx context.Context, db *gorm.DB) error {
	const name = "idx_account_roles_live_uq"

	var invalid bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
			WHERE c.relname = ? AND NOT i.indisvalid
		)`, name).Scan(&invalid).Error; err != nil {
		return err
	}
	if invalid {
		if err := db.WithContext(ctx).Exec(`DROP INDEX CONCURRENTLY IF EXISTS ` + name).Error; err != nil {
			return err
		}
	}

	if err := db.WithContext(ctx).Exec(`
		UPDATE account_roles ar
		SET deleted_at = EXTRACT(EPOCH FROM now())::bigint, updated_at = now()
		FROM (
			SELECT id, ROW_NUMBER() OVER (
				PARTITION BY account_id, role_id
				ORDER BY (org_id IS NULL), created_at, id
			) AS rn
			FROM account_roles
			WHERE deleted_at = 0
		) d
		WHERE ar.id = d.id AND d.rn > 1
	`).Error; err != nil {
		return err
	}

	return db.WithContext(ctx).Exec(`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS ` + name + `
ON account_roles (account_id, role_id) WHERE deleted_at = 0`).Error
}
