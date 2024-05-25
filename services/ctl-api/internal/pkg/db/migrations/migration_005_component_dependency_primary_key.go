package migrations

import "context"

// NOTE(jm): this was landed originally without the delete cascade
func (a *Migrations) migration005ComponentDependencyPrimaryKey(ctx context.Context) error {
	sql := `
ALTER TABLE component_dependencies DROP COLUMN IF EXISTS id;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
