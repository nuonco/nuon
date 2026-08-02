package account

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (c *Client) CreateServiceAccount(ctx context.Context, svcAcctID, name string) (*app.Account, error) {
	email := ServiceAccountEmail(svcAcctID)
	acct := app.Account{
		Email:       email,
		Subject:     svcAcctID,
		Name:        name,
		AccountType: app.AccountTypeService,
	}
	res := c.db.WithContext(ctx).
		Create(&acct)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create account")
	}

	return &acct, nil
}

// EnsureServiceAccount returns the service account for svcAcctID, creating it if it
// does not exist yet. CreateServiceAccount is a bare insert that conflicts on the
// unique email, so callers that may run against entities predating service-account
// creation need this instead.
func (c *Client) EnsureServiceAccount(ctx context.Context, svcAcctID, name string) (*app.Account, error) {
	acct, err := c.FindAccount(ctx, ServiceAccountEmail(svcAcctID))
	if err == nil {
		return acct, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "unable to look up service account")
	}

	return c.CreateServiceAccount(ctx, svcAcctID, name)
}
