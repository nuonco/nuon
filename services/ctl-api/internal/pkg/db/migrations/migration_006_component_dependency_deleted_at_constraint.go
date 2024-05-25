package migrations

import "context"

// NOTE(jm): this was landed originally without the delete cascade
func (a *Migrations) migration006ComponentDependencyDeletedAtConstraint(ctx context.Context) error {
	sql := `
ALTER TABLE component_dependencies DROP CONSTRAINT IF EXISTS idx_component_dependencies_deleted_at;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
