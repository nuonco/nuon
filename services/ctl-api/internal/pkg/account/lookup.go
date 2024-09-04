package account

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func ServiceAccountEmail(id string) string {
	return fmt.Sprintf("%s@serviceaccount.nuon.co", id)
}

func (c *Client) FindAccount(ctx context.Context, emailOrSubject string) (*app.Account, error) {
	acct := app.Account{}
	res := c.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Where(app.Account{
			Email: emailOrSubject,
		}).
		Or(app.Account{
			Subject: emailOrSubject,
		}).
		First(&acct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find account %s: %w", emailOrSubject, res.Error)
	}

	return &acct, nil
}
