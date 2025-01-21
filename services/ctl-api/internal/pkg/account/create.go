package account

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func (m *Client) CreateAccount(ctx context.Context, email, subject string) (*app.Account, error) {
	acct := app.Account{
		Email:       email,
		Subject:     subject,
		AccountType: app.AccountTypeAuth0,
	}

	if err := m.db.WithContext(ctx).
		Create(&acct).Error; err != nil {
		return nil, fmt.Errorf("unable to create account: %w", err)
	}

	ctx = cctx.SetAccountContext(ctx, &acct)
	m.analyticsClient.Identify(ctx)
	return &acct, nil
}
