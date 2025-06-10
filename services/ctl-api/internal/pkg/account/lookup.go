package account

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func ServiceAccountEmail(id string) string {
	return fmt.Sprintf("%s@serviceaccount.nuon.co", id)
}

func (c *Client) FindAccount(ctx context.Context, emailOrSubjectOrID string) (*app.Account, error) {
	acct := app.Account{}
	res := c.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Where("email = ?", emailOrSubjectOrID).
		Or(app.Account{
			Subject: emailOrSubjectOrID,
		}).
		Or(app.Account{
			Subject: emailOrSubjectOrID,
		}).
		First(&acct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find account %s: %w", emailOrSubjectOrID, res.Error)
	}

	return &acct, nil
}
