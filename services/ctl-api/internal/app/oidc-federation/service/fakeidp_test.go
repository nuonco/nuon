package service

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const testKeyID = "test-key"

// fakeIDP is a minimal OIDC issuer: a discovery document, a JWKS, and an RSA
// key that signs tokens.
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
		json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                idp.server.URL,
			"jwks_uri":                              idp.server.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
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

func (idp *fakeIDP) issuer() string { return idp.server.URL }

// signToken signs claims with the IdP key, filling iss/aud/exp/iat when absent.
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

func toJSONSerialization(t *testing.T, compact string, extra map[string]string) string {
	t.Helper()

	parts := strings.Split(compact, ".")
	require.Len(t, parts, 3)

	fields := map[string]string{
		"protected": parts[0],
		"payload":   parts[1],
		"signature": parts[2],
	}
	for k, v := range extra {
		fields[k] = v
	}

	out, err := json.Marshal(fields)
	require.NoError(t, err)
	return string(out)
}
