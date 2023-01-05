package middleware

import (
	"context"
	"fmt"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/pkg/errors"
)

type CustomClaim struct {
	ExternalID   string `json:"sub"`
	ShouldReject bool   `json:"shouldReject,omitempty"`
}

func (c CustomClaim) String() string {
	return fmt.Sprintf("%#v\n", c)
}

func (c *CustomClaim) Validate(context.Context) error {
	if c.ShouldReject {
		return errors.New("should reject was set to true")
	}

	if c.ExternalID == "" {
		return errors.New("subject must be set")
	}

	return nil
}

func Jwt(authIssuerURL, authAudience string) (*jwtmiddleware.JWTMiddleware, error) {
	issuerURL, err := url.Parse(authIssuerURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse issuer url")
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute) //nolint: gomnd

	customClaims := func() validator.CustomClaims {
		return &CustomClaim{}
	}

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		// NOTE(jm): in stage this must be "https://nuon.us.auth0.com/userinfo"
		// NOTE(jm): we can't pass more than one audience in
		// https://github.com/auth0/go-jwt-middleware/issues/148
		[]string{authAudience},
		validator.WithAllowedClockSkew(time.Minute),
		validator.WithCustomClaims(customClaims),
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not instantiate new jwt validator")
	}

	middleware := jwtmiddleware.New(jwtValidator.ValidateToken)
	return middleware, nil
}
