package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/install_inputs_view_v1.sql
var installInputsViewV1 string

func (a *Migrations) migration066InstallInputsViewV1(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(installInputsViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
