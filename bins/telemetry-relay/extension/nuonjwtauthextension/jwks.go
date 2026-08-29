package nuonjwtauthextension

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	maxJWKSSize           = 64 * 1024
	maxJWKSKeys           = 32
	jwksHTTPTimeout       = 5 * time.Second
	maxJWKSStaleness      = time.Hour
	unknownKeyRetryWindow = 5 * time.Second
)

var errJWKSUnavailable = errors.New("telemetry verification keys are unavailable")

type jsonWebKeySet struct {
	Keys []jsonWebKey `json:"keys"`
}

type jsonWebKey struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

type keyCache struct {
	url    string
	client *http.Client
	now    func() time.Time

	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastSuccess time.Time
	lastUnknown time.Time
	refreshMu   sync.Mutex
}

func newKeyCache(url string) *keyCache {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   jwksHTTPTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: jwksHTTPTimeout,
	}
	return &keyCache{
		url: url,
		client: &http.Client{
			Timeout:   jwksHTTPTimeout,
			Transport: transport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		now:  time.Now,
		keys: make(map[string]*rsa.PublicKey),
	}
}

func (c *keyCache) key(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	now := c.now()
	key, lastSuccess := c.cached(keyID)
	if key != nil && now.Sub(lastSuccess) <= maxJWKSStaleness {
		return key, nil
	}

	if err := c.refresh(ctx, true); err != nil {
		key, lastSuccess = c.cached(keyID)
		if key != nil && now.Sub(lastSuccess) <= maxJWKSStaleness {
			return key, nil
		}
		return nil, errJWKSUnavailable
	}
	key, lastSuccess = c.cached(keyID)
	if key == nil || now.Sub(lastSuccess) > maxJWKSStaleness {
		return nil, errJWKSUnavailable
	}
	return key, nil
}

func (c *keyCache) cached(keyID string) (*rsa.PublicKey, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.keys[keyID], c.lastSuccess
}

func (c *keyCache) refresh(ctx context.Context, throttle bool) error {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	now := c.now()
	c.mu.RLock()
	lastUnknown := c.lastUnknown
	c.mu.RUnlock()
	if throttle && !lastUnknown.IsZero() && now.Sub(lastUnknown) < unknownKeyRetryWindow {
		return nil
	}

	if throttle {
		c.mu.Lock()
		c.lastUnknown = now
		c.mu.Unlock()
	}

	keys, err := c.fetch(ctx)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.keys = keys
	c.lastSuccess = now
	c.mu.Unlock()
	return nil
}

func (c *keyCache) fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, errJWKSUnavailable
	}
	req.Header.Set("Accept", "application/json")
	response, err := c.client.Do(req)
	if err != nil {
		return nil, errJWKSUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errJWKSUnavailable
	}

	contents, err := io.ReadAll(io.LimitReader(response.Body, maxJWKSSize+1))
	if err != nil || len(contents) > maxJWKSSize {
		return nil, errJWKSUnavailable
	}
	keys, err := parseJWKS(contents)
	if err != nil {
		return nil, errJWKSUnavailable
	}
	return keys, nil
}

func parseJWKS(contents []byte) (map[string]*rsa.PublicKey, error) {
	var set jsonWebKeySet
	if err := json.Unmarshal(contents, &set); err != nil {
		return nil, fmt.Errorf("decode JWKS: %w", err)
	}
	if len(set.Keys) == 0 || len(set.Keys) > maxJWKSKeys {
		return nil, fmt.Errorf("JWKS must contain between one and %d keys", maxJWKSKeys)
	}

	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, key := range set.Keys {
		if key.KeyID == "" || key.KeyType != "RSA" || key.Use != "sig" || key.Algorithm != "RS256" {
			return nil, fmt.Errorf("JWKS contains an unsupported key")
		}
		if _, exists := keys[key.KeyID]; exists {
			return nil, fmt.Errorf("JWKS contains duplicate key IDs")
		}

		modulus, err := decodeJWKInteger(key.Modulus)
		if err != nil || modulus.Sign() <= 0 || modulus.BitLen() < 2048 {
			return nil, fmt.Errorf("JWKS contains an invalid RSA modulus")
		}
		exponent, err := decodeJWKInteger(key.Exponent)
		if err != nil || !exponent.IsInt64() || exponent.Int64() < 3 || exponent.Int64() > int64(^uint32(0)>>1) || exponent.Int64()%2 == 0 {
			return nil, fmt.Errorf("JWKS contains an invalid RSA exponent")
		}
		keys[key.KeyID] = &rsa.PublicKey{N: modulus, E: int(exponent.Int64())}
	}
	return keys, nil
}

func decodeJWKInteger(value string) (*big.Int, error) {
	if value == "" {
		return nil, errors.New("JWK integer is empty")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 {
		return nil, errors.New("JWK integer is invalid")
	}
	return new(big.Int).SetBytes(decoded), nil
}

func (c *keyCache) close() {
	c.client.CloseIdleConnections()
}
