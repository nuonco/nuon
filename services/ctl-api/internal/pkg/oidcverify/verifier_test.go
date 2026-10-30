package oidcverify

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const testKeyID = "test-key"

// fakeIDP is a minimal OIDC issuer: discovery document, JWKS, and an RSA key
// that signs tokens.
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
			"id_token_signing_alg_values_supported": []string{"RS256", "ES256"},
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

func TestVerify(t *testing.T) {
	idp := newFakeIDP(t)
	v := NewVerifier(func() bool { return true })
	ctx := context.Background()

	t.Run("valid token returns verified claims", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{
			"sub":              "repo:acme/app:ref:refs/heads/main",
			"repository_owner": "acme",
		})
		claims, err := v.Verify(ctx, token, idp.issuer(), "nuon-test")
		require.NoError(t, err)
		require.Equal(t, "repo:acme/app:ref:refs/heads/main", claims["sub"])
		require.Equal(t, "acme", claims["repository_owner"])
	})

	t.Run("wrong audience", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{"aud": "someone-else"})
		_, err := v.Verify(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
	})

	t.Run("wrong issuer claim", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{"iss": "https://evil.example.com"})
		_, err := v.Verify(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{
			"iat": time.Now().Add(-10 * time.Minute).Unix(),
			"exp": time.Now().Add(-time.Minute).Unix(),
		})
		_, err := v.Verify(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
	})

	t.Run("issued far in the future", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{
			"iat": time.Now().Add(10 * time.Minute).Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		_, err := v.Verify(ctx, token, idp.issuer(), "nuon-test")
		require.ErrorContains(t, err, "in the future")
	})

	t.Run("signed by a different key", func(t *testing.T) {
		other := newFakeIDP(t)
		token := other.signToken(t, jwt.MapClaims{"iss": idp.issuer()})
		_, err := v.Verify(ctx, token, idp.issuer(), "nuon-test")
		require.Error(t, err)
	})

	t.Run("discovery issuer mismatch", func(t *testing.T) {
		_, err := NewVerifier(func() bool { return true }).Verify(ctx, "a.b.c", "http://127.0.0.1:1", "aud")
		require.Error(t, err)
	})
}

func TestVerifyRejectsAlgorithmsDiscoveryAdvertises(t *testing.T) {
	idp := newFakeIDP(t)
	v := NewVerifier(func() bool { return true })

	// The discovery document advertises ES256, but verification is pinned to RS256.
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES256","kid":"test-key"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"` + idp.issuer() + `","aud":"nuon-test","sub":"x"}`))
	token := header + "." + payload + ".signature"

	_, err := v.Verify(context.Background(), token, idp.issuer(), "nuon-test")
	require.Error(t, err)
}

func TestVerifyIgnoresClaimsOutsideTheSignedPayload(t *testing.T) {
	idp := newFakeIDP(t)
	v := NewVerifier(func() bool { return true })

	token := idp.signToken(t, jwt.MapClaims{"sub": "signed-subject", "oid": "signed-oid"})
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)

	forged := base64.RawURLEncoding.EncodeToString([]byte(
		`{"iss":"` + idp.issuer() + `","aud":"nuon-test","sub":"forged-subject","oid":"forged-oid","exp":` +
			strconv.FormatInt(time.Now().Add(5*time.Minute).Unix(), 10) + `}`,
	))
	wrapped, err := json.Marshal(map[string]string{
		"protected": parts[0],
		"payload":   parts[1],
		"signature": parts[2],
		"fake":      "." + forged + ".",
	})
	require.NoError(t, err)

	claims, err := v.Verify(context.Background(), string(wrapped), idp.issuer(), "nuon-test")
	if err != nil {
		return
	}
	require.Equal(t, "signed-subject", claims["sub"])
	require.Equal(t, "signed-oid", claims["oid"])
}

func TestDiscoverCachesProvider(t *testing.T) {
	idp := newFakeIDP(t)
	v := NewVerifier(func() bool { return true })
	ctx := context.Background()

	require.NoError(t, v.Discover(ctx, idp.issuer()))

	var hits int
	idp.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
	})

	// A second discovery and a verification both reuse the cached provider, so
	// the now-failing issuer is never contacted again for discovery.
	require.NoError(t, v.Discover(ctx, idp.issuer()))
	token := idp.signToken(t, jwt.MapClaims{"sub": "x"})
	_, err := v.Verify(ctx, token, idp.issuer(), "nuon-test")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "discovery")
}

func TestUnverifiedIssuer(t *testing.T) {
	idp := newFakeIDP(t)

	t.Run("extracts issuer", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{})
		issuer, err := UnverifiedIssuer(token)
		require.NoError(t, err)
		require.Equal(t, idp.issuer(), issuer)
	})

	t.Run("rejects non-jwt input", func(t *testing.T) {
		_, err := UnverifiedIssuer("not-a-jwt")
		require.Error(t, err)
	})

	t.Run("rejects flattened jws", func(t *testing.T) {
		_, err := UnverifiedIssuer(`{"payload":"e30","signatures":[{"protected":"e30","signature":"e30"}]}`)
		require.Error(t, err)
	})

	t.Run("rejects missing iss", func(t *testing.T) {
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"x"}`))
		_, err := UnverifiedIssuer("aGVhZGVy." + payload + ".c2ln")
		require.Error(t, err)
	})

	t.Run("rejects wrong segment counts", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{})
		parts := strings.Split(token, ".")

		cases := []string{
			token + ".extra",
			parts[0] + "." + parts[1],
		}
		for _, tc := range cases {
			_, err := UnverifiedIssuer(tc)
			require.Error(t, err, tc)
		}
	})

	t.Run("rejects invalid base64 payload", func(t *testing.T) {
		token := idp.signToken(t, jwt.MapClaims{})
		parts := strings.Split(token, ".")

		_, err := UnverifiedIssuer(parts[0] + ".not!valid!base64." + parts[2])
		require.Error(t, err)
	})
}
