package migrations

import "context"

func (m *Migrations) migration053UserTokensLegacyIndexes(ctx context.Context) error {
	sql := `ALTER TABLE user_tokens DROP CONSTRAINT idx_user_tokens_subject;`
	if res := m.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
