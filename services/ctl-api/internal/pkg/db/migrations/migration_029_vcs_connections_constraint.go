package migrations

import (
	"context"
)

func (a *Migrations) migration029VcsConnectionsConstraint(ctx context.Context) error {
	sql := `
ALTER TABLE vcs_connection_commits DROP CONSTRAINT IF EXISTS fk_vcs_connection_commits_org;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil

}
