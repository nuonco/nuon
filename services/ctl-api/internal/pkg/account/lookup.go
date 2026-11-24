package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func ServiceAccountEmail(id string) string {
	return fmt.Sprintf("%s@serviceaccount.nuon.co", id)
}

func (c *Client) FindAccount(ctx context.Context, emailOrSubjectOrID string) (*app.Account, error) {
	query := c.db.
		Distinct().
		Where("email = ?", strings.ToLower(emailOrSubjectOrID)).
		Or(app.Account{
			Subject: emailOrSubjectOrID,
		}).
		Or(app.Account{
			ID: emailOrSubjectOrID,
		})
	return c.loadAccount(ctx, query, fmt.Sprintf("unable to find account %s", emailOrSubjectOrID))
}
