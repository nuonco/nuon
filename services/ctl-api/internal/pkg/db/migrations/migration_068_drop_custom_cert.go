package migrations

import (
	"context"
)

func (a *Migrations) migration068DropCustomCert(ctx context.Context) error {
	sql := `
ALTER TABLE orgs DROP COLUMN IF EXISTS custom_cert
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
