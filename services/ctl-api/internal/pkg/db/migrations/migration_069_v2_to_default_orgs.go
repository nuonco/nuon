package migrations

import (
	"context"
)

func (a *Migrations) migration069V2ToDefaultOrgs(ctx context.Context) error {
	sql := `
UPDATE orgs SET org_type='default' WHERE org_type='v2'
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
