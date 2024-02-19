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

	return nil
}
