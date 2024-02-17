package migrations

import "context"

func (a *Migrations) migration022RemoveDuplicateUserTokens(ctx context.Context) error {
	sql := `

DELETE FROM user_tokens
WHERE id IN
    (SELECT id
    FROM
	(SELECT id,
	 ROW_NUMBER() OVER( PARTITION BY subject
	ORDER BY id DESC ) AS row_num
	FROM user_tokens ) t
	WHERE t.row_num > 1 );
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
