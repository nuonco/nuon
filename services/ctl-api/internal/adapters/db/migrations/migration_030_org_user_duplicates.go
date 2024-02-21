package migrations

import (
	"context"
)

func (a *Migrations) migration030OrgUserDuplicates(ctx context.Context) error {
	sql := `
DELETE FROM user_orgs
WHERE id IN
    (SELECT id
    FROM
	(SELECT id,
	 ROW_NUMBER() OVER( PARTITION BY org_id, user_id
	ORDER BY id DESC ) AS row_num
	FROM user_orgs ) t
	WHERE t.row_num > 1 );
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	sql = `
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_org on user_orgs USING btree (deleted_at, user_id, org_id)
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
