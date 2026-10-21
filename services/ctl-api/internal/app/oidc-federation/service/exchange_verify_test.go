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
	"strings"
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
		_, err := svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test")
		require.NoError(t, err)
	})

	t.Run("wrong audience", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{"aud": "someone-else"})
		_, err := svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
	})

	t.Run("wrong issuer claim", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{"iss": "https://evil.example.com"})
		_, err := svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{
			"iat": time.Now().Add(-10 * time.Minute).Unix(),
			"exp": time.Now().Add(-5 * time.Minute).Unix(),
		})
		_, err := svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
	})

	t.Run("token signed by a different key", func(t *testing.T) {
		otherIDP := newFakeIDP(t)
		token := otherIDP.signToken(t, jwt.MapClaims{"iss": idp.issuer()})
		_, err := svc.verifyOIDCToken(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
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

func TestVerifyOIDCTokenReturnsVerifiedClaims(t *testing.T) {
	idp := newFakeIDP(t)
	svc := newVerifyTestService()

	token := idp.signToken(t, jwt.MapClaims{
		"sub":              "repo:acme/app:ref:refs/heads/main",
		"repository_owner": "acme",
	})

	claims, err := svc.verifyOIDCToken(context.Background(), token, idp.issuer(), "nuon-test")
	require.NoError(t, err)
	require.Equal(t, "repo:acme/app:ref:refs/heads/main", claims.Sub)
	owner, _ := claims.Lookup("repository_owner")
	require.Equal(t, "acme", owner)
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

func TestCompactJWSOnly(t *testing.T) {
	idp := newFakeIDP(t)
	svc := newVerifyTestService()
	ctx := context.Background()

	compact := idp.signToken(t, jwt.MapClaims{"sub": "repo:attacker/app:ref:refs/heads/main"})
	parts := strings.Split(compact, ".")

	unsigned, err := json.Marshal(map[string]any{
		"iss": idp.issuer(),
		"aud": "nuon-test",
		"sub": "repo:acme/app:ref:refs/heads/main",
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	})
	require.NoError(t, err)
	dotted := "." + base64.RawURLEncoding.EncodeToString(unsigned) + "."

	tests := map[string]struct {
		token   string
		wantErr bool
	}{
		"compact token": {
			token: compact,
		},
		"flattened JSON serialization": {
			token:   toJSONSerialization(t, compact, nil),
			wantErr: true,
		},
		"flattened JSON with dotted unknown property": {
			token:   toJSONSerialization(t, compact, map[string]string{"extra": dotted}),
			wantErr: true,
		},
		"too many parts": {
			token:   compact + ".extra",
			wantErr: true,
		},
		"too few parts": {
			token:   parts[0] + "." + parts[1],
			wantErr: true,
		},
		"empty signature": {
			token:   parts[0] + "." + parts[1] + ".",
			wantErr: true,
		},
		"non-base64url characters": {
			token:   parts[0] + "." + parts[1] + "." + parts[2] + "+/=",
			wantErr: true,
		},
		"trailing newline": {
			token:   compact + "\n",
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, issuerErr := parseUnverifiedIssuer(tc.token)
			claims, verifyErr := svc.verifyOIDCToken(ctx, tc.token, idp.issuer(), "nuon-test")

			if tc.wantErr {
				require.Error(t, issuerErr)
				require.Error(t, verifyErr)
				require.Zero(t, claims)
				return
			}

			require.NoError(t, issuerErr)
			require.NoError(t, verifyErr)
			require.Equal(t, "repo:attacker/app:ref:refs/heads/main", claims.Sub)
		})
	}
}
