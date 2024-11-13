package account

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Client) CreateServiceAccount(ctx context.Context, svcAcctID string) (*app.Account, error) {
	email := ServiceAccountEmail(svcAcctID)
	acct := app.Account{
		Email:       email,
		Subject:     svcAcctID,
		AccountType: app.AccountTypeService,
	}
	res := c.db.WithContext(ctx).
		Create(&acct)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create account")
	}

	return &acct, nil
}
