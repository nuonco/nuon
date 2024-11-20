package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/actions_view_v1.sql
var actionsViewV1 string

func (a *Migrations) migration073ActionsView(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(actionsViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
