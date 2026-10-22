package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
)

const defaultFailureTTL = 30 * time.Second

var ErrInvalidJWTFormat = errors.New("invalid JWT format")

func CompactJWSParts(token string) ([]string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidJWTFormat
	}
	for _, part := range parts {
		if part == "" || strings.IndexFunc(part, isNotBase64URL) >= 0 {
			return nil, ErrInvalidJWTFormat
		}
	}
	return parts, nil
}

func isNotBase64URL(r rune) bool {
	return !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_')
}

type VerifierOptions struct {
	HTTPClient           *http.Client
	AllowInsecureIssuers bool
	FailureTTL           time.Duration
}

type Verifier struct {
	opts    VerifierOptions
	baseCtx context.Context

	mu      sync.Mutex
	entries map[string]*providerEntry
}

type providerEntry struct {
	mu        sync.Mutex
	provider  *gooidc.Provider
	err       error
	failedAt  time.Time
	attempted bool
}

func NewVerifier(opts VerifierOptions) *Verifier {
	if opts.FailureTTL <= 0 {
		opts.FailureTTL = defaultFailureTTL
	}

	baseCtx := context.Background()
	if opts.HTTPClient != nil {
		baseCtx = gooidc.ClientContext(baseCtx, opts.HTTPClient)
	}

	return &Verifier{
		opts:    opts,
		baseCtx: baseCtx,
		entries: make(map[string]*providerEntry),
	}
}

func (v *Verifier) Verify(ctx context.Context, token, issuer, audience string) (Claims, error) {
	if _, err := CompactJWSParts(token); err != nil {
		return Claims{}, err
	}

	provider, err := v.provider(issuer)
	if err != nil {
		return Claims{}, err
	}

	idToken, err := provider.Verifier(&gooidc.Config{
		ClientID:             audience,
		SupportedSigningAlgs: []string{gooidc.RS256},
	}).Verify(ctx, token)
	if err != nil {
		return Claims{}, fmt.Errorf("JWT validation failed: %w", err)
	}

	var raw map[string]any
	if err := idToken.Claims(&raw); err != nil {
		return Claims{}, fmt.Errorf("failed to read JWT claims: %w", err)
	}

	claims, err := DecodeClaims(raw)
	if err != nil {
		return Claims{}, err
	}
	if claims.Iat == 0 {
		return Claims{}, errors.New("JWT missing iat claim")
	}

	return claims, nil
}

func (v *Verifier) provider(issuer string) (*gooidc.Provider, error) {
	v.mu.Lock()
	entry, ok := v.entries[issuer]
	if !ok {
		entry = &providerEntry{}
		v.entries[issuer] = entry
	}
	v.mu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.provider != nil {
		return entry.provider, nil
	}
	if entry.attempted && time.Since(entry.failedAt) < v.opts.FailureTTL {
		return nil, entry.err
	}

	provider, err := v.discover(issuer)
	if err != nil {
		entry.attempted = true
		entry.failedAt = time.Now()
		entry.err = err
		return nil, err
	}

	entry.provider = provider
	entry.attempted = false
	entry.err = nil
	return provider, nil
}

func (v *Verifier) discover(issuer string) (*gooidc.Provider, error) {
	provider, err := gooidc.NewProvider(v.baseCtx, issuer)
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery failed: %w", err)
	}

	var doc struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := provider.Claims(&doc); err != nil {
		return nil, fmt.Errorf("failed to read OIDC discovery document: %w", err)
	}

	jwksURL, err := url.Parse(doc.JWKSURI)
	if err != nil {
		return nil, fmt.Errorf("invalid jwks_uri: %w", err)
	}
	switch jwksURL.Scheme {
	case "https":
	case "http":
		if !v.opts.AllowInsecureIssuers {
			return nil, errors.New("jwks_uri must use https")
		}
	default:
		return nil, errors.New("jwks_uri must use https")
	}

	return provider, nil
}
