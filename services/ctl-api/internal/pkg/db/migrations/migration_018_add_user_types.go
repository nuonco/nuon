package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration018AddUserTypes(ctx context.Context) error {
	var orgs []*app.Token
	res := a.db.WithContext(ctx).
		Find(&orgs)
	if res.Error != nil {
		return res.Error
	}

	for _, org := range orgs {
		a.l.Info("adding user token type")
		res := a.db.WithContext(ctx).
			Model(&app.Token{
				ID: org.ID,
			})
			// Updates(app.Token{TokenType: tokenTyp})
		if res.Error != nil {
			return fmt.Errorf("unable to add token type: %w", res.Error)
		}
	}

	return nil
}
