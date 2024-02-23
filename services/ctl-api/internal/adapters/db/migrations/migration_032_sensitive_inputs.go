package migrations

import (
	"context"
)

func (a *Migrations) migration033SensitiveInputs(ctx context.Context) error {
	sql := `
update app_inputs set sensitive=false where sensitive is null;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil

}
