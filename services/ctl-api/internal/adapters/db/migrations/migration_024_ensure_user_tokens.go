package migrations

import (
	"context"
	"errors"

	"time"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (a *Migrations) migration024EnsureUserTokens(ctx context.Context) error {
	var orgs []*app.Org
	res := a.db.WithContext(ctx).
		Find(&orgs)
	if res.Error != nil {
		return res.Error
	}

	for _, org := range orgs {
		var userToken *app.UserToken
		res = a.db.WithContext(ctx).
			Where("subject = ?", org.CreatedByID).
			First(&userToken)
		if res.Error == nil {
			continue
		}
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return res.Error
		}

		res = a.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&app.UserToken{
				Token:     generics.GetFakeObj[string](),
				TokenType: app.TokenTypeAuth0,
				Subject:   org.CreatedByID,
				Issuer:    "auth0",
				ExpiresAt: time.Now().Add(time.Hour * 24),
				Email:     "",
			})
		if res.Error != nil {
			return res.Error
		}
	}

	return nil
}
