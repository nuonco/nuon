package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

func (m *middleware) fetchUserToken(ctx context.Context, token string) (*app.UserToken, error) {
	var userToken app.UserToken
	res := m.db.
		WithContext(ctx).
		Where(&app.UserToken{
			Token: token,
		}).
		First(&userToken)

	// no error found
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if res.Error != nil {
		return nil, fmt.Errorf("error occurred querying user tokens: %w", res.Error)
	}

	// make sure this is not an expired token
	if time.Now().After(userToken.ExpiresAt) {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("token is expired"),
			Description: "Please get a new token from the Nuon dashboard",
		}
	}

	return &userToken, nil
}

func (m *middleware) saveUserToken(ctx context.Context, token string, claims *validator.ValidatedClaims) (*app.UserToken, error) {
	userToken := app.UserToken{
		Token:     token,
		Subject:   claims.RegisteredClaims.Subject,
		ExpiresAt: time.Unix(claims.RegisteredClaims.Expiry, 0),
		IssuedAt:  time.Unix(claims.RegisteredClaims.IssuedAt, 0),
		Issuer:    claims.RegisteredClaims.Issuer,
	}

	res := m.db.WithContext(ctx).Create(&userToken)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to save user token: %w", res.Error)
	}

	return &userToken, nil
}
