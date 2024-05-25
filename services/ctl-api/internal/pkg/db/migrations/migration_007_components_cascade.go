package migrations

import "context"

func (a *Migrations) migration007ComponentDependencyCascade(ctx context.Context) error {
	sql := `
ALTER TABLE component_dependencies DROP CONSTRAINT IF EXISTS fk_component_dependencies_component;
ALTER TABLE component_dependencies ADD CONSTRAINT fk_component_dependencies_component
	FOREIGN KEY (component_id)
	REFERENCES components(id)
	ON DELETE CASCADE;

ALTER TABLE component_dependencies DROP CONSTRAINT IF EXISTS fk_component_dependencies_dependencies;
ALTER TABLE component_dependencies ADD CONSTRAINT fk_component_dependencies_dependencies
	FOREIGN KEY (dependency_id)
	REFERENCES components(id)
	ON DELETE CASCADE;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
