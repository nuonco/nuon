package service

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
)

const jwksCacheDuration = 5 * time.Minute

// jwksProviderCache caches JWKS providers per issuer to avoid repeated OIDC
// discovery. Issuer URLs always come from stored trust policies, never from
// presented tokens.
type jwksProviderCache struct {
	mu        sync.RWMutex
	providers map[string]*jwks.CachingProvider
}

func newJWKSProviderCache() *jwksProviderCache {
	return &jwksProviderCache{
		providers: make(map[string]*jwks.CachingProvider),
	}
}

func (p *jwksProviderCache) getProvider(issuer string) (*jwks.CachingProvider, error) {
	p.mu.RLock()
	provider, ok := p.providers[issuer]
	p.mu.RUnlock()
	if ok {
		return provider, nil
	}

	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer URL: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if provider, ok := p.providers[issuer]; ok {
		return provider, nil
	}

	provider = jwks.NewCachingProvider(issuerURL, jwksCacheDuration)
	p.providers[issuer] = provider
	return provider, nil
}
