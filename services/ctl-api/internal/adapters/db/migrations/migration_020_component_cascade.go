package migrations

import (
	"context"
)

func (a *Migrations) migration020InstallComponentCascades(ctx context.Context) error {
	sql := `
ALTER TABLE install_components DROP CONSTRAINT IF EXISTS fk_install_components_component;
ALTER TABLE install_components ADD CONSTRAINT fk_component_dependencies_component
	FOREIGN KEY (component_id)
	REFERENCES components(id)
	ON DELETE CASCADE;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
