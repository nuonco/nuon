package account

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TODO(jm): this entire file should probably live in `pkg/account`
func (m *Client) FetchAccount(ctx context.Context, acctID string) (*app.Account, error) {
	return m.loadAccount(ctx, m.db.Where(&app.Account{ID: acctID}, "ID"), fmt.Sprintf("unable to fetch account %s", acctID))
}

func (m *Client) loadAccount(ctx context.Context, query *gorm.DB, lookupDescription string) (*app.Account, error) {
	acct := app.Account{}
	res := query.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		First(&acct)
	if res.Error != nil {
		return nil, fmt.Errorf("%s: %w", lookupDescription, res.Error)
	}

	authority, err := m.authzClient.LoadAccountPermissions(ctx, &acct)
	if err != nil {
		return nil, fmt.Errorf("unable to load account permissions: %w", err)
	}
	acct.Authority = authority

	return &acct, nil
}
