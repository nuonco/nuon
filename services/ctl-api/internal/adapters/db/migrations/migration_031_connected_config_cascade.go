package migrations

import "context"

func (a *Migrations) migration031ConnectedVCSConfigCascadeConstraint(ctx context.Context) error {
	sql := `
ALTER TABLE connected_github_vcs_configs DROP CONSTRAINT IF EXISTS fk_connected_github_vcs_configs_org;
ALTER TABLE connected_github_vcs_configs ADD CONSTRAINT fk_connected_github_vcs_configs_org
	FOREIGN KEY (org_id)
	REFERENCES orgs(id)
	ON DELETE CASCADE;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
