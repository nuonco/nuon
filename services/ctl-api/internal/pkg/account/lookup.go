package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func ServiceAccountEmail(id string) string {
	return fmt.Sprintf("%s@serviceaccount.nuon.co", id)
}

func (c *Client) FindAccount(ctx context.Context, emailOrSubjectOrID string) (*app.Account, error) {
	acct := app.Account{}
	res := c.db.WithContext(ctx).
		Distinct().
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Where("email = ?", strings.ToLower(emailOrSubjectOrID)).
		Or(app.Account{
			Subject: emailOrSubjectOrID,
		}).
		Or(app.Account{
			ID: emailOrSubjectOrID,
		}).
		First(&acct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find account %s: %w", emailOrSubjectOrID, res.Error)
	}

	authority, err := authz.New(authz.Params{DB: c.db}).LoadAccountPermissions(ctx, &acct)
	if err != nil {
		return nil, fmt.Errorf("unable to load account permissions: %w", err)
	}
	acct.Authority = authority

	return &acct, nil
}
