package auth

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"go.uber.org/zap"
)

func (m *middleware) validateToken(ctx context.Context, token string) (*validator.ValidatedClaims, error) {
	m.l.Debug("validating token",
		zap.String("audience", m.cfg.Auth0Audience),
		zap.String("issuer-url", m.cfg.Auth0IssuerURL),
		zap.String("token", token),
	)

	issuerURL, err := url.Parse(m.cfg.Auth0IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse issuer url: %w", err)
	}

	customClaimsFn := func() validator.CustomClaims {
		return &customClaims{}
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	tokenValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		m.cfg.Auth0IssuerURL,
		[]string{m.cfg.Auth0Audience},
		validator.WithAllowedClockSkew(time.Minute),
		validator.WithCustomClaims(customClaimsFn),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create validator: %w", err)
	}

	validToken, err := tokenValidator.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("unable to validate token: %w", err)
	}

	claims, ok := validToken.(*validator.ValidatedClaims)
	if !ok {
		return nil, fmt.Errorf("invalid response from token validator: %w", err)
	}

	return claims, nil
}
