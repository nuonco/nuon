// Package workloadjwt verifies cloud-issued workload identity JWTs.
//
// Callers supply the issuer and audience; they are never read out of the presented
// token, so an unauthenticated caller cannot steer which key set is trusted.
package workloadjwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/pkg/jwtcompact"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcverify"
)

type Request struct {
	Token    string
	Issuer   string
	Audience string
}

func (r Request) validate() error {
	switch {
	case strings.TrimSpace(r.Token) == "":
		return errors.New("token is required")
	case strings.TrimSpace(r.Issuer) == "":
		return errors.New("issuer is required")
	case strings.TrimSpace(r.Audience) == "":
		return errors.New("audience is required")
	}

	issuerURL, err := url.Parse(r.Issuer)
	if err != nil {
		return fmt.Errorf("invalid issuer url: %w", err)
	}
	if issuerURL.Scheme != "https" || issuerURL.Host == "" {
		return fmt.Errorf("issuer %q must be an https url", r.Issuer)
	}

	return nil
}

type Verifier struct {
	oidc *oidcverify.Verifier
}

func NewVerifier() *Verifier {
	return &Verifier{oidc: oidcverify.NewVerifier(func() bool { return false })}
}

// Verify checks signature, issuer, audience and time claims, and returns the claims
// from the signed payload.
//
// A valid signature only establishes which cloud tenant minted the token, not that it is
// the right one, so callers must bind the returned claims to stored state.
func (v *Verifier) Verify(ctx context.Context, req Request) (map[string]any, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	return v.oidc.Verify(ctx, req.Token, req.Issuer, req.Audience)
}

// UnverifiedClaims decodes claims without checking the signature.
func UnverifiedClaims(token string) (map[string]any, error) {
	token = strings.TrimSpace(token)
	if err := RequireCompactJWT(token); err != nil {
		return nil, err
	}

	parts := strings.Split(token, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to decode jwt payload: %w", err)
	}

	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unable to parse jwt claims: %w", err)
	}

	return claims, nil
}

func RequireCompactJWT(token string) error {
	return jwtcompact.Require(token)
}

// StringClaim reads a string claim. A claim of any other type is treated as absent
// rather than coerced.
func StringClaim(claims map[string]any, name string) (string, bool) {
	raw, ok := claims[name]
	if !ok {
		return "", false
	}

	value, ok := raw.(string)
	if !ok || value == "" {
		return "", false
	}

	return value, true
}
