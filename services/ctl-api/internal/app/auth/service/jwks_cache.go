package service

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
)

// JWKSCache manages JWKS providers for multiple cloud providers with caching
type JWKSCache struct {
	providers map[string]*jwks.CachingProvider
	mu        sync.RWMutex
	ttl       time.Duration
}

// NewJWKSCache creates a new JWKS cache with the specified TTL
func NewJWKSCache(ttl time.Duration) *JWKSCache {
	return &JWKSCache{
		providers: make(map[string]*jwks.CachingProvider),
		ttl:       ttl,
	}
}

// GetProvider returns a JWKS provider for the given issuer URL, creating it if necessary
func (c *JWKSCache) GetProvider(ctx context.Context, issuerURL string) (*jwks.CachingProvider, error) {
	// Try to get existing provider with read lock
	c.mu.RLock()
	provider, exists := c.providers[issuerURL]
	c.mu.RUnlock()

	if exists {
		return provider, nil
	}

	// Acquire write lock to create new provider
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have created it)
	if provider, exists := c.providers[issuerURL]; exists {
		return provider, nil
	}

	// Validate and parse issuer URL
	parsedURL, err := url.Parse(issuerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer URL %q: %w", issuerURL, err)
	}

	// Create new caching provider
	provider = jwks.NewCachingProvider(parsedURL, c.ttl)
	c.providers[issuerURL] = provider

	return provider, nil
}

// Clear removes all cached providers (useful for testing)
func (c *JWKSCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.providers = make(map[string]*jwks.CachingProvider)
}
