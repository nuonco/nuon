package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const testKeyID = "test-key"

// fakeIDP is a minimal OIDC issuer: it serves an OIDC discovery document and
// a JWKS for a generated RSA key, and signs tokens with that key.
type fakeIDP struct {
	server *httptest.Server
	key    *rsa.PrivateKey
}

func newFakeIDP(t *testing.T) *fakeIDP {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	idp := &fakeIDP{key: key}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"issuer":   idp.server.URL,
			"jwks_uri": idp.server.URL + "/jwks",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{
				{
					"kty": "RSA",
					"kid": testKeyID,
					"use": "sig",
					"alg": "RS256",
					"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
				},
			},
		})
	})

	idp.server = httptest.NewServer(mux)
	t.Cleanup(idp.server.Close)

	return idp
}

func (idp *fakeIDP) issuer() string {
	return idp.server.URL
}

// signToken signs claims with the IdP key, defaulting iss/aud/exp/iat when
// not provided.
func (idp *fakeIDP) signToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()

	if _, ok := claims["iss"]; !ok {
		claims["iss"] = idp.issuer()
	}
	if _, ok := claims["aud"]; !ok {
		claims["aud"] = "nuon-test"
	}
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = time.Now().Unix()
	}
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = time.Now().Add(5 * time.Minute).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID

	signed, err := token.SignedString(idp.key)
	require.NoError(t, err)
	return signed
}

func newVerifyTestService() *service {
	return &service{jwks: newJWKSProviderCache()}
}

func TestVerifyOIDCToken(t *testing.T) {
	idp := newFakeIDP(t)
	svc := newVerifyTestService()
	ctx := context.Background()

	t.Run("valid token", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{"sub": "repo:acme/app:ref:refs/heads/main"})
		require.NoError(t, svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test"))
	})

	t.Run("wrong audience", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{"aud": "someone-else"})
		require.Error(t, svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test"))
	})

	t.Run("wrong issuer claim", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{"iss": "https://evil.example.com"})
		require.Error(t, svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test"))
	})

	t.Run("expired token", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{
			"iat": time.Now().Add(-10 * time.Minute).Unix(),
			"exp": time.Now().Add(-5 * time.Minute).Unix(),
		})
		require.Error(t, svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test"))
	})

	t.Run("token signed by a different key", func(t *testing.T) {
		otherIDP := newFakeIDP(t)
		token := otherIDP.signToken(t, jwt.MapClaims{"iss": idp.issuer()})
		require.Error(t, svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test"))
	})
}

func TestParseUnverifiedIssuer(t *testing.T) {
	idp := newFakeIDP(t)

	t.Run("extracts issuer", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{})
		issuer, err := parseUnverifiedIssuer(token)
		require.NoError(t, err)
		require.Equal(t, idp.issuer(), issuer)
	})

	t.Run("rejects non-JWT input", func(t *testing.T) {
		_, err := parseUnverifiedIssuer("not-a-jwt")
		require.Error(t, err)
	})

	t.Run("rejects missing iss claim", func(t *testing.T) {
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"x"}`))
		_, err := parseUnverifiedIssuer("aGVhZGVy." + payload + ".c2ln")
		require.Error(t, err)
	})
}

func TestParseTokenClaims(t *testing.T) {
	idp := newFakeIDP(t)

	token := idp.signToken(t, jwt.MapClaims{
		"sub":              "repo:acme/app:ref:refs/heads/main",
		"repository_owner": "acme",
	})

	claims, err := parseTokenClaims(token)
	require.NoError(t, err)
	require.Equal(t, "repo:acme/app:ref:refs/heads/main", claims["sub"])
	require.Equal(t, "acme", claims["repository_owner"])
}
