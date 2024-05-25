package migrations

import "context"

func (a *Migrations) migration001ExampleSQL(ctx context.Context) error {
	sql := `
DROP TABLE IF EXISTS does_not_exist;
`

	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	a.l.Info("example migration - sql")
	return nil
}
