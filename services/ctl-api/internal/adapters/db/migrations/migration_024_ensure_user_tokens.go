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
	var deletedOrgs []*app.Org
	res := a.db.Unscoped().WithContext(ctx).
		Preload("Apps").
		Preload("Apps.Installs").
		Preload("Apps.Components").
		Find(&deletedOrgs)
	if res.Error != nil {
		return res.Error
	}

	var orgs []*app.Org
	res = a.db.WithContext(ctx).
		Preload("Apps").
		Preload("Apps.Installs").
		Preload("Apps.Components").
		Find(&orgs)
	if res.Error != nil {
		return res.Error
	}

	allOrgs := append(orgs, deletedOrgs...)

	userIDs := make([]string, 0)
	for _, org := range allOrgs {
		userIDs = append(userIDs, org.CreatedByID)

		for _, app := range org.Apps {
			userIDs = append(userIDs, app.CreatedByID)

			for _, install := range app.Installs {
				userIDs = append(userIDs, install.CreatedByID)
			}

			for _, comp := range app.Components {
				userIDs = append(userIDs, comp.CreatedByID)
			}
		}
	}

	for _, userID := range userIDs {
		var userToken *app.UserToken
		res = a.db.WithContext(ctx).
			Where("subject = ?", userID).
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
				CreatedByID: userID,
				Token:       generics.GetFakeObj[string](),
				TokenType:   app.TokenTypeAuth0,
				Subject:     userID,
				Issuer:      "auth0",
				ExpiresAt:   time.Now().Add(time.Hour * 24),
				Email:       "",
			})
		if res.Error != nil {
			return res.Error
		}
	}

	return nil
}
