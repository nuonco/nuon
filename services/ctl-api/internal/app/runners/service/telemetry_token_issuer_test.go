package service

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func telemetryTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func telemetryTestJWK(key *rsa.PrivateKey, keyID string, includePrivate bool) telemetryJSONWebKey {
	jwk := telemetryJSONWebKey{
		KeyType:   "RSA",
		KeyID:     keyID,
		Use:       "sig",
		Algorithm: jwt.SigningMethodRS256.Alg(),
		Modulus:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		Exponent:  encodeTelemetryJWKInteger(int64(key.E)),
	}
	if includePrivate {
		jwk.D = base64.RawURLEncoding.EncodeToString(key.D.Bytes())
		jwk.P = base64.RawURLEncoding.EncodeToString(key.Primes[0].Bytes())
		jwk.Q = base64.RawURLEncoding.EncodeToString(key.Primes[1].Bytes())
	}
	return jwk
}

func telemetryTestJWKS(t *testing.T, keys ...telemetryJSONWebKey) string {
	t.Helper()

	contents, err := json.Marshal(telemetryJSONWebKeySet{Keys: keys})
	require.NoError(t, err)
	return string(contents)
}

func newTelemetryTestTokenIssuer(t *testing.T) (*telemetryTokenIssuer, time.Time) {
	t.Helper()

	key := telemetryTestRSAKey(t)
	issuer, err := newTelemetryTokenIssuer(&internal.Config{
		PublicAPIURL: "https://ctl.example.com/",
		TelemetryJWKS: telemetryTestJWKS(t,
			telemetryTestJWK(key, "telemetry-key-1", true),
		),
	})
	require.NoError(t, err)

	now := time.Date(2026, time.August, 28, 20, 0, 0, 0, time.UTC)
	issuer.now = func() time.Time { return now }
	return issuer, now
}

func TestTelemetryTokenIssuerIssuesScopedAccessToken(t *testing.T) {
	issuer, now := newTelemetryTestTokenIssuer(t)
	principal := telemetryRunnerPrincipal{
		OrgID:     "org-test",
		AppID:     "app-test",
		InstallID: "install-test",
		RunnerID:  "runner-test",
	}

	raw, err := issuer.issue(principal)
	require.NoError(t, err)

	claims := &telemetryAccessTokenClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return &issuer.privateKey.PublicKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithIssuer(issuer.issuer), jwt.WithAudience(telemetryTokenAudience), jwt.WithTimeFunc(func() time.Time { return now }))
	require.NoError(t, err)
	require.True(t, token.Valid)
	require.Equal(t, "at+jwt", token.Header["typ"])
	require.Equal(t, "telemetry-key-1", token.Header["kid"])
	require.Equal(t, telemetryTokenScope, claims.Scope)
	require.Equal(t, principal.RunnerID, claims.ClientID)
	require.Equal(t, principal.OrgID, claims.OrgID)
	require.Equal(t, principal.AppID, claims.AppID)
	require.Equal(t, principal.InstallID, claims.InstallID)
	require.Equal(t, principal.RunnerID, claims.RunnerID)
	require.Equal(t, "org:org-test:install:install-test:runner:runner-test", claims.Subject)
	require.True(t, now.Equal(claims.IssuedAt.Time))
	require.True(t, now.Equal(claims.NotBefore.Time))
	require.True(t, now.Add(telemetryTokenLifetime).Equal(claims.ExpiresAt.Time))
	require.NotEmpty(t, claims.ID)

	parts := strings.Split(raw, ".")
	require.Len(t, parts, 3)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	var wireClaims map[string]any
	require.NoError(t, json.Unmarshal(payload, &wireClaims))
	require.Equal(t, []any{telemetryTokenAudience}, wireClaims["aud"])
}

func TestTelemetryTokenIssuerPublishesOnlyPublicKeyMaterial(t *testing.T) {
	issuer, _ := newTelemetryTestTokenIssuer(t)

	keys := issuer.publicJWKS()
	require.Len(t, keys.Keys, 1)
	require.Equal(t, TelemetryJSONWebKey{
		KeyType:   "RSA",
		KeyID:     "telemetry-key-1",
		Use:       "sig",
		Algorithm: "RS256",
		Modulus:   base64.RawURLEncoding.EncodeToString(issuer.privateKey.N.Bytes()),
		Exponent:  encodeTelemetryJWKInteger(int64(issuer.privateKey.E)),
	}, keys.Keys[0])

	encoded, err := json.Marshal(keys)
	require.NoError(t, err)
	var published map[string][]map[string]any
	require.NoError(t, json.Unmarshal(encoded, &published))
	require.NotContains(t, published["keys"][0], "d")
	require.NotContains(t, published["keys"][0], "p")
	require.NotContains(t, published["keys"][0], "q")
}

func TestTelemetryTokenIssuerRequiresOnePrivateSigningKey(t *testing.T) {
	key := telemetryTestRSAKey(t)

	t.Run("public keys only", func(t *testing.T) {
		_, err := newTelemetryTokenIssuer(&internal.Config{
			PublicAPIURL: "https://ctl.example.com",
			TelemetryJWKS: telemetryTestJWKS(t,
				telemetryTestJWK(key, "public-key", false),
			),
		})
		require.ErrorContains(t, err, "does not contain a private signing key")
	})

	t.Run("multiple private keys", func(t *testing.T) {
		_, err := newTelemetryTokenIssuer(&internal.Config{
			PublicAPIURL: "https://ctl.example.com",
			TelemetryJWKS: telemetryTestJWKS(t,
				telemetryTestJWK(key, "key-1", true),
				telemetryTestJWK(key, "key-2", true),
			),
		})
		require.ErrorContains(t, err, "exactly one private signing key")
	})
}

func TestCreateTelemetryAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	issuer, _ := newTelemetryTestTokenIssuer(t)
	svc := &service{
		db:                   setupTelemetryRunnerPrincipalDB(t),
		telemetryTokenIssuer: issuer,
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/telemetry/access-token", nil)
	cctx.SetAccountGinContext(ctx, telemetryRunnerTestAccount())
	cctx.SetOrgIDGinContext(ctx, telemetryTestOrgID)

	svc.CreateTelemetryAccessToken(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	require.Equal(t, "no-cache", recorder.Header().Get("Pragma"))
	var response CreateTelemetryAccessTokenResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "Bearer", response.TokenType)
	require.Equal(t, int64(600), response.ExpiresIn)
	require.NotEmpty(t, response.AccessToken)
}

func TestGetTelemetryJWKS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	issuer, _ := newTelemetryTestTokenIssuer(t)
	svc := &service{telemetryTokenIssuer: issuer}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)

	svc.GetTelemetryJWKS(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "public, max-age=300", recorder.Header().Get("Cache-Control"))
	var response TelemetryJSONWebKeySet
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, issuer.publicJWKS(), response)
}

func TestTelemetryEndpointsUnavailableWithoutIssuer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &service{}

	for _, test := range []struct {
		name   string
		method string
		path   string
		invoke func(*gin.Context)
	}{
		{
			name:   "access token",
			method: http.MethodPost,
			path:   "/v1/telemetry/access-token",
			invoke: svc.CreateTelemetryAccessToken,
		},
		{
			name:   "public keys",
			method: http.MethodGet,
			path:   "/.well-known/jwks.json",
			invoke: svc.GetTelemetryJWKS,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(test.method, test.path, nil)

			test.invoke(ctx)

			require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
		})
	}
}
