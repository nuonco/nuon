package migrations

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *Migrations) migration055UserTokenAccountIDs(ctx context.Context) error {
	var accounts []app.Account
	res := m.db.WithContext(ctx).
		Find(&accounts)
	if res.Error != nil {
		return res.Error
	}

	byEmail := make(map[string]app.Account)
	for _, acct := range accounts {
		byEmail[acct.Email] = acct
	}

	var userTokens []app.Token
	res = m.db.WithContext(ctx).
		Find(&userTokens)
	if res.Error != nil {
		return res.Error
	}

	for _, userToken := range userTokens {
		if userToken.AccountID != "" {
			continue
		}

		res = m.db.WithContext(ctx).
			Model(app.Token{
				ID: userToken.ID,
			}).
			Updates(app.Token{
				AccountID: byEmail[userToken.Email].ID,
			})
		if res.Error != nil {
			return res.Error
		}
	}

	return nil
}
