package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration018AddUserTypes(ctx context.Context) error {
	var orgs []*app.UserToken
	res := a.db.WithContext(ctx).
		Find(&orgs)
	if res.Error != nil {
		return res.Error
	}

	for _, org := range orgs {
		tokenTyp := app.TokenTypeAuth0
		if strings.HasPrefix(org.Subject, "can") {
			tokenTyp = app.TokenTypeCanary
		}
		if strings.HasPrefix(org.Subject, "int") {
			tokenTyp = app.TokenTypeIntegration
		}

		a.l.Info("adding user token type")
		res := a.db.WithContext(ctx).
			Model(&app.UserToken{
				ID: org.ID,
			}).
			Updates(app.UserToken{TokenType: tokenTyp})
		if res.Error != nil {
			return fmt.Errorf("unable to add token type: %w", res.Error)
		}
	}

	// hard delete any soft-deleted orgs
	var deletedUserTokens []app.UserToken
	res = a.db.WithContext(ctx).Unscoped().Find(&deletedUserTokens)
	if res.Error != nil {
		return fmt.Errorf("unable to find deleted user tokens: %w", res.Error)
	}

	if len(deletedUserTokens) < 1 {
		return nil
	}

	res = a.db.WithContext(ctx).Unscoped().Delete(&deletedUserTokens)
	if res.Error != nil {
		return fmt.Errorf("unable to hard delete deleted user tokens: %w", res.Error)
	}

	return nil
}
