package migrations

import "context"

func (a *Migrations) migration019OrgAndUserTypesNotNullable(ctx context.Context) error {
	sql := `
ALTER TABLE orgs ALTER COLUMN org_type SET NOT NULL;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	sql = `
ALTER TABLE user_tokens ALTER COLUMN token_type SET NOT NULL;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
