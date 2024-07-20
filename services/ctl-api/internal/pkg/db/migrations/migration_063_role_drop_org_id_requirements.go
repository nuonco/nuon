package migrations

import "context"

func (a *Migrations) migration063RoleDropOrgIDRequirements(ctx context.Context) error {
	// org IDs are no longer required for roles and policies
	sql := `
ALTER TABLE roles ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE policies ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE account_roles ALTER COLUMN org_id DROP NOT NULL;

DROP INDEX IF EXISTS idx_org_role_policy;
DROP INDEX IF EXISTS idx_role;

ALTER TABLE roles DROP CONSTRAINT IF EXISTS fk_orgs_roles;
ALTER TABLE policies DROP CONSTRAINT IF EXISTS fk_orgs_policies;
ALTER TABLE account_roles DROP CONSTRAINT IF EXISTS fk_orgs_account_roles;
`

	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	a.l.Info("example migration - sql")
	return nil
}
