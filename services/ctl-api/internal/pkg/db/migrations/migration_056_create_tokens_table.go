package migrations

import "context"

// create a tokens table, so we can migrate the tokens table in place
func (m *Migrations) migration056CreateTokensTable(ctx context.Context) error {
	sql := `CREATE TABLE tokens AS
  SELECT * FROM user_tokens
`
	if res := m.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}
	return nil
}
