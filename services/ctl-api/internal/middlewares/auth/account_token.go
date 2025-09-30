package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
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

type customClaims struct {
	Email string `json:"email"`
}

func (c customClaims) Validate(ctx context.Context) error {
	return nil
}

func (m *middleware) saveAccountToken(ctx context.Context, token string, claims *validator.ValidatedClaims) (*app.Token, error) {
	customClaims, ok := claims.CustomClaims.(*customClaims)
	if !ok {
		return nil, fmt.Errorf("unable to get custom claims")
	}

	if customClaims.Email == "" {
		return nil, fmt.Errorf("email is empty in custom claims")
	}

	acct, err := m.acctClient.FindAccount(ctx, customClaims.Email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unable to get account: %w", err)
		}

		// Determine appropriate user journey based on signup type
		var pendingInvite app.OrgInvite
		inviteErr := m.db.WithContext(ctx).Where(&app.OrgInvite{
			Email:  customClaims.Email,
			Status: app.OrgInviteStatusPending,
		}).First(&pendingInvite).Error

		var userJourneys app.UserJourneys
		if inviteErr == nil {
			// Found pending invite - create account without journey tracking
			userJourneys = account.NoUserJourneys()
		} else {
			// No pending invite - self-signup user, check deployment configuration
			if m.cfg.EvaluationJourneyEnabled {
				// Multi-tenant deployment: Enable evaluation journey
				userJourneys = account.DefaultEvaluationJourney()
			} else {
				// BYOC deployment: Skip evaluation journey for clean first-run experience
				userJourneys = account.NoUserJourneys()
			}
		}

		acct, err = m.acctClient.CreateAccount(ctx, customClaims.Email, claims.RegisteredClaims.Subject, userJourneys)
		if err != nil {
			return nil, fmt.Errorf("unable to create account: %w", err)
		}
	}

	acctToken := app.Token{
		Token:       token,
		TokenType:   app.TokenTypeAuth0,
		ExpiresAt:   time.Unix(claims.RegisteredClaims.Expiry, 0),
		IssuedAt:    time.Unix(claims.RegisteredClaims.IssuedAt, 0),
		Issuer:      claims.RegisteredClaims.Issuer,
		CreatedByID: claims.RegisteredClaims.Subject,
		AccountID:   acct.ID,
	}

	res := m.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token"}},
			UpdateAll: true,
		}).
		Create(&acctToken)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to save user token: %w", res.Error)
	}

	return &acctToken, nil
}
