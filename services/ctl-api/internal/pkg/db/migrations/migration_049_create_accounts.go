package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *Migrations) migration049CreateAccounts(ctx context.Context) error {
	var tokens []app.Token

	res := m.db.
		Unscoped().
		WithContext(ctx).
		Find(&tokens)
	if res.Error != nil {
		return fmt.Errorf("unable to get tokens: %w", res.Error)
	}

	lookup := make(map[[2]string]struct{})
	rows := make([]app.Account, 0)
	for _, token := range tokens {
		key := [2]string{token.Email, token.Subject}
		_, ok := lookup[key]
		if ok {
			continue
		}

		var typ app.AccountType
		switch token.TokenType {
		case app.TokenTypeAuth0:
			typ = app.AccountTypeAuth0
		case app.TokenTypeAdmin:
			typ = app.AccountTypeService
		case app.TokenTypeStatic:
			typ = app.AccountTypeService
		case app.TokenTypeIntegration:
			typ = app.AccountTypeIntegration
		case app.TokenTypeCanary:
			typ = app.AccountTypeCanary
		}

		// for legacy (pre-accounts), the subject was used as the created by id
		rows = append(rows, app.Account{
			ID:          token.Subject,
			Email:       token.Email,
			Subject:     token.Subject,
			AccountType: typ,
		})
	}

	if len(rows) < 1 {
		return nil
	}

	// create all the accounts
	res = m.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoNothing: true,
			Columns: []clause.Column{
				{
					Name: "deleted_at",
				},
				{
					Name: "email",
				},
				{
					Name: "subject",
				},
			},
		}).
		Create(rows)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
