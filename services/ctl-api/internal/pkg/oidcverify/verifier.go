package oidcverify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/sync/singleflight"

	"github.com/nuonco/nuon/pkg/jwtcompact"
)

const (
	// RS256 is the only algorithm accepted. Discovery may advertise more, but
	// widening that set is a policy decision, not something an issuer gets to make.
	signingAlgorithm = oidc.RS256

	// issuedAtSkew bounds how far in the future an iat claim may be. go-oidc
	// checks exp strictly and does not expose a skew setting.
	issuedAtSkew = time.Minute

	// maxCachedIssuers bounds discovery state if issuers are ever passed in
	// unconstrained. Eviction is oldest-first.
	maxCachedIssuers = 256
)

// Verifier verifies OIDC ID tokens.
//
// Providers are cached per issuer: oidc.Provider owns a key set that caches
// JWKS and refreshes on an unknown kid, so constructing one per request would
// repeat discovery and throw away rotated keys.
type Verifier struct {
	// allowInsecure reports whether issuers off the public internet are
	// permitted. Read per request so a config change takes effect without
	// rebuilding the verifier.
	allowInsecure func() bool

	mu        sync.Mutex
	providers map[string]*oidc.Provider
	inserted  []string
	group     singleflight.Group
}

// NewVerifier returns a verifier. allowInsecure permits issuers that are not on
// the public internet, for tests and local development.
func NewVerifier(allowInsecure func() bool) *Verifier {
	return &Verifier{
		allowInsecure: allowInsecure,
		providers:     map[string]*oidc.Provider{},
	}
}

// Discover confirms the issuer serves a discovery document whose issuer matches,
// and caches the provider. Call it when an issuer is configured so a typo fails
// the configuration request instead of the first token exchange.
func (v *Verifier) Discover(ctx context.Context, issuer string) error {
	_, err := v.provider(ctx, issuer)
	return err
}

// Verify checks the token's signature, issuer, audience, and time claims against
// the given issuer and audience, and returns the verified claims.
//
// issuer and audience must come from stored configuration. The unverified iss
// claim is only safe for choosing which configuration to load.
func (v *Verifier) Verify(ctx context.Context, rawToken, issuer, audience string) (map[string]any, error) {
	provider, err := v.provider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	idToken, err := provider.Verifier(&oidc.Config{
		ClientID:             audience,
		SupportedSigningAlgs: []string{signingAlgorithm},
	}).Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	if skew := idToken.IssuedAt.Sub(time.Now()); skew > issuedAtSkew {
		return nil, fmt.Errorf("token issued %s in the future", skew.Round(time.Second))
	}

	claims := map[string]any{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("token claims: %w", err)
	}

	return claims, nil
}

func (v *Verifier) provider(ctx context.Context, issuer string) (*oidc.Provider, error) {
	v.mu.Lock()
	if provider, ok := v.providers[issuer]; ok {
		v.mu.Unlock()
		return provider, nil
	}
	v.mu.Unlock()

	discovered, err, _ := v.group.Do(issuer, func() (any, error) {
		// Re-check under the singleflight key: the outer check above raced
		// with another caller that may have already finished discovery for
		// this issuer between that check and this one starting.
		v.mu.Lock()
		if provider, ok := v.providers[issuer]; ok {
			v.mu.Unlock()
			return provider, nil
		}
		v.mu.Unlock()

		discoveryCtx := oidc.ClientContext(ctx, newHTTPClient(v.allowInsecure()))
		provider, err := oidc.NewProvider(discoveryCtx, issuer)
		if err != nil {
			return nil, fmt.Errorf("oidc discovery for %s failed: %w", issuer, err)
		}

		// No re-check needed here: singleflight runs this function at most
		// once at a time per issuer, and this is the only writer to
		// v.providers, so nothing could have inserted issuer since the check
		// above.
		v.mu.Lock()
		defer v.mu.Unlock()
		if len(v.inserted) >= maxCachedIssuers {
			oldest := v.inserted[0]
			v.inserted = v.inserted[1:]
			delete(v.providers, oldest)
		}
		v.providers[issuer] = provider
		v.inserted = append(v.inserted, issuer)
		return provider, nil
	})
	if err != nil {
		return nil, err
	}

	return discovered.(*oidc.Provider), nil
}

// UnverifiedIssuer extracts the iss claim without checking the signature. It
// exists to select stored configuration; authorization must use Verify.
// Only compact JWTs are accepted so a flattened JWS cannot steer discovery.
func UnverifiedIssuer(rawToken string) (string, error) {
	if err := jwtcompact.Require(rawToken); err != nil {
		return "", err
	}

	parts := strings.Split(strings.TrimSpace(rawToken), ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("malformed jwt payload: %w", err)
	}

	var claims struct {
		Issuer string `json:"iss"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("malformed jwt claims: %w", err)
	}
	if claims.Issuer == "" {
		return "", errors.New("jwt is missing iss")
	}

	return claims.Issuer, nil
}
