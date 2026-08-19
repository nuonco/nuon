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
	"sync"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

const (
	jwksCacheTTL     = 5 * time.Minute
	allowedClockSkew = time.Minute

	// Bounds memory and outbound discovery if a caller ever passes an issuer it did
	// not fully constrain.
	maxCachedIssuers = 256
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

	return nil
}

type Verifier struct {
	mu        sync.Mutex
	providers map[string]*jwks.CachingProvider
	inserted  []string
}

func NewVerifier() *Verifier {
	return &Verifier{providers: map[string]*jwks.CachingProvider{}}
}

// Verify checks signature, issuer, audience and time claims.
//
// A valid signature only establishes which cloud tenant minted the token, not that it is
// the right one, so callers must bind the returned claims to stored state.
func (v *Verifier) Verify(ctx context.Context, req Request) (map[string]any, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	provider, err := v.provider(req.Issuer)
	if err != nil {
		return nil, err
	}

	tokenValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		req.Issuer,
		[]string{req.Audience},
		validator.WithAllowedClockSkew(allowedClockSkew),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build token validator: %w", err)
	}

	if _, err := tokenValidator.ValidateToken(ctx, req.Token); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Decoded rather than read off the validator, which only surfaces registered claims
	// unless a concrete type is registered up front. Safe: the signature over this
	// payload is already checked.
	claims, err := decodeClaims(req.Token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (v *Verifier) provider(issuer string) (*jwks.CachingProvider, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if provider, ok := v.providers[issuer]; ok {
		return provider, nil
	}

	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer url: %w", err)
	}
	if issuerURL.Scheme != "https" || issuerURL.Host == "" {
		return nil, fmt.Errorf("issuer %q must be an https url", issuer)
	}

	if len(v.inserted) >= maxCachedIssuers {
		oldest := v.inserted[0]
		v.inserted = v.inserted[1:]
		delete(v.providers, oldest)
	}

	provider := jwks.NewCachingProvider(issuerURL, jwksCacheTTL)
	v.providers[issuer] = provider
	v.inserted = append(v.inserted, issuer)

	return provider, nil
}

// UnverifiedClaims decodes claims without checking the signature. Only for selecting
// which key set to verify against; never for authorization.
func UnverifiedClaims(token string) (map[string]any, error) {
	return decodeClaims(token)
}

func decodeClaims(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed jwt")
	}

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
