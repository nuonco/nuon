package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Several identity providers can now share a provider_type, so provider_type stops being an
// identifier and identity_provider_id becomes one.
//
// The old idx_provider_type was declared with a malformed gorm tag and landed as
// UNIQUE (deleted_at, org_id) with no provider_type column at all, so it actually allowed only a
// single global provider row of any type.
//
// The backfill uses 'default-' || provider_type rather than a bare sentinel so it cannot collide:
// the indexes it replaces were unique on (account_id, provider_type) and (provider_type, sub).
func (m *Migrations) Migration129IdentityProvidersMultiPerType(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		DO $$
		DECLARE
			con record;
		BEGIN
			FOR con IN
				SELECT conname FROM pg_constraint
				WHERE conrelid = 'account_identities'::regclass
					AND contype = 'f'
					AND confrelid = 'identity_providers'::regclass
			LOOP
				EXECUTE format('ALTER TABLE account_identities DROP CONSTRAINT %I', con.conname);
			END LOOP;
		END $$;

		UPDATE account_identities
		SET identity_provider_id = 'default-' || provider_type
		WHERE identity_provider_id IS NULL;

		ALTER TABLE account_identities ALTER COLUMN identity_provider_id SET NOT NULL;

		DROP INDEX IF EXISTS idx_account_identity_account_provider;
		DROP INDEX IF EXISTS idx_account_identity_provider_sub;
		DROP INDEX IF EXISTS idx_account_identities_provider_type_sub;

		DROP INDEX IF EXISTS idx_provider_type;
	`).Error
}
