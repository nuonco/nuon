package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (m *middleware) fetchAccountToken(ctx context.Context, token string) (*app.Token, error) {
	var userToken app.Token
	res := m.db.
		WithContext(ctx).
		Where(&app.Token{
			Token: token,
		}).
		First(&userToken)

	// no error found
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if res.Error != nil {
		return nil, fmt.Errorf("error occurred querying account tokens: %w", res.Error)
	}

	// make sure this is not an expired token
	if time.Now().After(userToken.ExpiresAt) {
		return nil, stderr.ErrAuthentication{
			Err:         fmt.Errorf("token is expired"),
			Description: "Please get a new token from the Nuon dashboard",
		}
	}

	return &userToken, nil
}
