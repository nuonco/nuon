package migrations

import "context"

func (a *Migrations) migration016InputCascades(ctx context.Context) error {
	sql := `
ALTER TABLE install_inputs DROP CONSTRAINT IF EXISTS fk_install_inputs_app_input_config;
ALTER TABLE install_inputs ADD CONSTRAINT fk_install_inputs_app_input_config
	FOREIGN KEY (app_input_config_id)
	REFERENCES app_input_configs(id)
	ON DELETE CASCADE;

ALTER TABLE installs DROP CONSTRAINT IF EXISTS fk_installs_app_input_config;
ALTER TABLE installs ADD CONSTRAINT fk_installs_app_input_config
	FOREIGN KEY (app_input_config_id)
	REFERENCES app_input_configs(id)
	ON DELETE CASCADE;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
