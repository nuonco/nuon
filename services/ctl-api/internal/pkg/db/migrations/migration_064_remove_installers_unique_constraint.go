package migrations

import "context"

func (a *Migrations) migration064RemoveInstallerConstraint(ctx context.Context) error {
	sql := `ALTER TABLE installers DROP CONSTRAINT IF EXISTS idx_installers_org_id`

	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
