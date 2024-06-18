package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSyncTokenRequest struct {
	AccountID string
}

type CreateSyncTokenResponse struct {
	Token string
}

func (a *Activities) CreateSyncToken(ctx context.Context, req CreateSyncTokenRequest) (*CreateSyncTokenResponse, error) {
	token := app.Token{
		CreatedByID: req.AccountID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeCanary,
		ExpiresAt:   time.Now().Add(time.Hour),
		IssuedAt:    time.Now(),
		Issuer:      fmt.Sprintf("config-sync-%s", req.AccountID),
		AccountID:   req.AccountID,
	}

	res := a.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create sync token: %w", res.Error)
	}
	return &CreateSyncTokenResponse{
		Token: token.Token,
	}, nil
}
