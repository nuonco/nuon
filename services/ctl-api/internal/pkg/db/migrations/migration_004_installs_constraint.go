package migrations

import "context"

// NOTE(jm): this was landed originally without the delete cascade
func (a *Migrations) migration004InstallsCascadeInputs(ctx context.Context) error {
	sql := `
ALTER TABLE install_inputs DROP CONSTRAINT fk_installs_install_inputs;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	sql = `
ALTER TABLE install_inputs ADD CONSTRAINT fk_installs_install_inputs
	FOREIGN KEY (install_id)
	REFERENCES installs(id)
	ON DELETE CASCADE
	;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	a.l.Info("example migration - sql")
	return nil
}
