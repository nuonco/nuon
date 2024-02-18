package migrations

import "context"

func (a *Migrations) migration023UserTokensUniqueConstraint(ctx context.Context) error {
	sql := `
ALTER TABLE user_tokens DROP CONSTRAINT IF EXISTS idx_user_tokens_subject;

ALTER TABLE user_tokens ADD CONSTRAINT idx_user_tokens_subject UNIQUE("subject");
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
