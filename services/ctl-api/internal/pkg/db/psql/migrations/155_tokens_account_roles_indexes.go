package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (m *Migrations) Migration155TokensAccountRolesIndexes(ctx context.Context, db *gorm.DB) error {
	if err := ensureValidUniqueIndex(ctx, db, "tokens_pkey", `
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS tokens_pkey
ON tokens (id)`); err != nil {
		return err
	}

	var hasPK bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conrelid = 'tokens'::regclass AND contype = 'p'
		)`).Scan(&hasPK).Error; err != nil {
		return fmt.Errorf("check tokens primary key: %w", err)
	}
	if !hasPK {
		if err := db.WithContext(ctx).Exec(`
ALTER TABLE tokens ADD CONSTRAINT tokens_pkey PRIMARY KEY USING INDEX tokens_pkey`).Error; err != nil {
			return fmt.Errorf("add tokens primary key: %w", err)
		}
	}

	if err := db.WithContext(ctx).Exec(`
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_roles_role_id
ON account_roles (role_id)`).Error; err != nil {
		return fmt.Errorf("create account_roles role_id index: %w", err)
	}

	if err := db.WithContext(ctx).Exec(`
ALTER TABLE account_roles DROP CONSTRAINT IF EXISTS fk_account_roles_account_role`).Error; err != nil {
		return fmt.Errorf("drop unused account_roles self-reference: %w", err)
	}

	return nil
}

func ensureValidUniqueIndex(ctx context.Context, db *gorm.DB, name, createSQL string) error {
	var invalid bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_index i
			JOIN pg_class c ON c.oid = i.indexrelid
			WHERE c.relname = ? AND NOT i.indisvalid
		)`, name).Scan(&invalid).Error; err != nil {
		return fmt.Errorf("check invalid index %s: %w", name, err)
	}
	if invalid {
		if err := db.WithContext(ctx).Exec(`DROP INDEX CONCURRENTLY IF EXISTS ` + name).Error; err != nil {
			return fmt.Errorf("drop invalid index %s: %w", name, err)
		}
	}
	if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
		return fmt.Errorf("create index %s: %w", name, err)
	}
	return nil
}
