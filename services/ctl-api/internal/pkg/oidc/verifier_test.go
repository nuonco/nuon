package oidc_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
)

const testAudience = "nuon-test"

type idpKey struct {
	kid string
	key *rsa.PrivateKey
}

type fakeIDP struct {
	server *httptest.Server

	mu            sync.Mutex
	keys          []idpKey
	issuerClaim   string
	jwksURIScheme string
	discoveryHits atomic.Int32
	failDiscovery atomic.Bool
}

func newKey(t *testing.T, kid string) idpKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return idpKey{kid: kid, key: key}
}

func startIDP(t *testing.T, tls bool) *fakeIDP {
	t.Helper()

	idp := &fakeIDP{keys: []idpKey{newKey(t, "key-1")}}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		idp.discoveryHits.Add(1)
		if idp.failDiscovery.Load() {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		idp.mu.Lock()
		defer idp.mu.Unlock()
		issuer := idp.server.URL
		if idp.issuerClaim != "" {
			issuer = idp.issuerClaim
		}
		jwksURI := idp.server.URL + "/jwks"
		if idp.jwksURIScheme != "" {
			jwksURI = idp.jwksURIScheme + "://" + strings.SplitN(idp.server.URL, "://", 2)[1] + "/jwks"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"issuer": issuer, "jwks_uri": jwksURI})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		idp.mu.Lock()
		defer idp.mu.Unlock()
		var keys []map[string]string
		for _, k := range idp.keys {
			keys = append(keys, map[string]string{
				"kty": "RSA",
				"kid": k.kid,
				"use": "sig",
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(k.key.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(k.key.PublicKey.E)).Bytes()),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"keys": keys})
	})

	if tls {
		idp.server = httptest.NewTLSServer(mux)
	} else {
		idp.server = httptest.NewServer(mux)
	}
	t.Cleanup(idp.server.Close)
	return idp
}

func (idp *fakeIDP) addKey(t *testing.T, kid string) idpKey {
	t.Helper()
	k := newKey(t, kid)
	idp.mu.Lock()
	idp.keys = append(idp.keys, k)
	idp.mu.Unlock()
	return k
}

func (idp *fakeIDP) sign(t *testing.T, k idpKey, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()
	if _, ok := claims["iss"]; !ok {
		claims["iss"] = idp.server.URL
	}
	if _, ok := claims["aud"]; !ok {
		claims["aud"] = testAudience
	}
	if v, ok := claims["iat"]; ok && v == nil {
		delete(claims, "iat")
	} else if !ok {
		claims["iat"] = time.Now().Unix()
	}
	if v, ok := claims["exp"]; ok && v == nil {
		delete(claims, "exp")
	} else if !ok {
		claims["exp"] = time.Now().Add(5 * time.Minute).Unix()
	}
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = k.kid
	signed, err := token.SignedString(k.key)
	require.NoError(t, err)
	return signed
}

func (idp *fakeIDP) signRS256(t *testing.T, claims jwt.MapClaims) string {
	return idp.sign(t, idp.keys[0], jwt.SigningMethodRS256, claims)
}

func newTestVerifier(client *http.Client, allowInsecure bool) *oidc.Verifier {
	return oidc.NewVerifier(oidc.VerifierOptions{
		HTTPClient:           client,
		AllowInsecureIssuers: allowInsecure,
	})
}

func TestVerifierVerify(t *testing.T) {
	now := time.Now()

	tests := map[string]struct {
		claims  jwt.MapClaims
		wantErr bool
	}{
		"valid token": {
			claims: jwt.MapClaims{"sub": "repo:acme/app:ref:refs/heads/main", "repository_owner": "acme"},
		},
		"missing exp": {
			claims:  jwt.MapClaims{"exp": nil},
			wantErr: true,
		},
		"missing iat": {
			claims:  jwt.MapClaims{"iat": nil},
			wantErr: true,
		},
		"expired": {
			claims: jwt.MapClaims{
				"iat": now.Add(-10 * time.Minute).Unix(),
				"exp": now.Add(-5 * time.Minute).Unix(),
			},
			wantErr: true,
		},
		"wrong audience": {
			claims:  jwt.MapClaims{"aud": "someone-else"},
			wantErr: true,
		},
		"wrong issuer claim": {
			claims:  jwt.MapClaims{"iss": "https://evil.example.com"},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			idp := startIDP(t, false)
			v := newTestVerifier(idp.server.Client(), true)

			token := idp.signRS256(t, tc.claims)
			claims, err := v.Verify(context.Background(), token, idp.server.URL, testAudience)

			if tc.wantErr {
				require.Error(t, err)
				assert.Zero(t, claims)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "repo:acme/app:ref:refs/heads/main", claims.Sub)
			assert.Equal(t, idp.server.URL, claims.Iss)
			assert.Equal(t, "acme", claims.Extra["repository_owner"])
		})
	}
}

func TestVerifierDiscoveryIssuerMismatch(t *testing.T) {
	idp := startIDP(t, false)
	idp.issuerClaim = "https://other.example.com"
	v := newTestVerifier(idp.server.Client(), true)

	token := idp.signRS256(t, jwt.MapClaims{})
	_, err := v.Verify(context.Background(), token, idp.server.URL, testAudience)
	require.Error(t, err)
}

func TestVerifierRejectsInsecureJWKSURI(t *testing.T) {
	idp := startIDP(t, true)
	idp.jwksURIScheme = "http"
	v := newTestVerifier(idp.server.Client(), false)

	token := idp.signRS256(t, jwt.MapClaims{})
	_, err := v.Verify(context.Background(), token, idp.server.URL, testAudience)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwks_uri must use https")
}

func TestVerifierKeyRotation(t *testing.T) {
	idp := startIDP(t, false)
	v := newTestVerifier(idp.server.Client(), true)
	ctx := context.Background()

	_, err := v.Verify(ctx, idp.signRS256(t, jwt.MapClaims{}), idp.server.URL, testAudience)
	require.NoError(t, err)

	rotated := idp.addKey(t, "key-2")
	token := idp.sign(t, rotated, jwt.SigningMethodRS256, jwt.MapClaims{})
	_, err = v.Verify(ctx, token, idp.server.URL, testAudience)
	require.NoError(t, err)
}

func TestVerifierRejectsMalformedAndWrongAlg(t *testing.T) {
	idp := startIDP(t, false)
	v := newTestVerifier(idp.server.Client(), true)

	compact := idp.signRS256(t, jwt.MapClaims{"sub": "x"})
	parts := strings.Split(compact, ".")
	jsonJWS, err := json.Marshal(map[string]string{
		"protected": parts[0],
		"payload":   parts[1],
		"signature": parts[2],
	})
	require.NoError(t, err)

	hs256 := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": idp.server.URL,
		"aud": testAudience,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	hs256.Header["kid"] = "key-1"
	hs256Token, err := hs256.SignedString([]byte("secret"))
	require.NoError(t, err)

	none := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"iss": idp.server.URL,
		"aud": testAudience,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	noneToken, err := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	tests := map[string]string{
		"json serialization": string(jsonJWS),
		"hs256":              hs256Token,
		"none":               noneToken,
		"empty":              "",
	}

	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			claims, err := v.Verify(context.Background(), token, idp.server.URL, testAudience)
			require.Error(t, err)
			assert.Zero(t, claims)
		})
	}
}

func TestVerifierCachesDiscoveryFailure(t *testing.T) {
	idp := startIDP(t, false)
	idp.failDiscovery.Store(true)
	v := oidc.NewVerifier(oidc.VerifierOptions{
		HTTPClient:           idp.server.Client(),
		AllowInsecureIssuers: true,
		FailureTTL:           time.Minute,
	})

	token := idp.signRS256(t, jwt.MapClaims{})
	for i := 0; i < 2; i++ {
		_, err := v.Verify(context.Background(), token, idp.server.URL, testAudience)
		require.Error(t, err)
	}
	assert.EqualValues(t, 1, idp.discoveryHits.Load())
}

func TestVerifierRetriesAfterFailureTTL(t *testing.T) {
	idp := startIDP(t, false)
	idp.failDiscovery.Store(true)
	v := oidc.NewVerifier(oidc.VerifierOptions{
		HTTPClient:           idp.server.Client(),
		AllowInsecureIssuers: true,
		FailureTTL:           10 * time.Millisecond,
	})

	token := idp.signRS256(t, jwt.MapClaims{})
	_, err := v.Verify(context.Background(), token, idp.server.URL, testAudience)
	require.Error(t, err)

	idp.failDiscovery.Store(false)
	time.Sleep(20 * time.Millisecond)

	_, err = v.Verify(context.Background(), token, idp.server.URL, testAudience)
	require.NoError(t, err)
}
