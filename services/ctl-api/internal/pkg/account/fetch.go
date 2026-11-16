package account

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func (m *Client) FetchAccountIdentity(ctx context.Context, acctID string) (*app.Account, error) {
	if acctID == "" {
		return nil, fmt.Errorf("account identity requires an account ID")
	}
	var acct app.Account
	if err := m.db.WithContext(ctx).Where(&app.Account{ID: acctID}).First(&acct).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch account identity: %w", err)
	}
	return &acct, nil
}

// TODO(jm): this entire file should probably live in `pkg/account`
func (m *Client) FetchAccount(ctx context.Context, acctID string) (*app.Account, error) {
	acct := app.Account{}
	res := m.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		First(&acct, "id = ?", acctID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to fetch account %s: %w", acctID, res.Error)
	}

	authority, err := authz.New(authz.Params{DB: m.db}).LoadLegacyAccountPermissions(ctx, &acct)
	if err != nil {
		return nil, fmt.Errorf("unable to load account permissions: %w", err)
	}
	acct.Authority = authority

	return &acct, nil
}
