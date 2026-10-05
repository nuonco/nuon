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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func telemetryTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

type telemetryTestPrivateJWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
	D         string `json:"d,omitempty"`
	P         string `json:"p,omitempty"`
	Q         string `json:"q,omitempty"`
	DP        string `json:"dp,omitempty"`
	DQ        string `json:"dq,omitempty"`
	QI        string `json:"qi,omitempty"`
}

func telemetryTestJWKInteger(value *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(value.Bytes())
}

func telemetryTestJWK(key *rsa.PrivateKey, keyID string, includePrivate bool) telemetryTestPrivateJWK {
	jwk := telemetryTestPrivateJWK{
		KeyType:   "RSA",
		KeyID:     keyID,
		Use:       "sig",
		Algorithm: jwt.SigningMethodRS256.Alg(),
		Modulus:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		Exponent:  telemetryTestJWKInteger(big.NewInt(int64(key.E))),
	}
	if includePrivate {
		jwk.D = base64.RawURLEncoding.EncodeToString(key.D.Bytes())
		jwk.P = base64.RawURLEncoding.EncodeToString(key.Primes[0].Bytes())
		jwk.Q = base64.RawURLEncoding.EncodeToString(key.Primes[1].Bytes())
	}
	return jwk
}

func telemetryTestJWKS(t *testing.T, keys ...telemetryTestPrivateJWK) string {
	t.Helper()

	contents, err := json.Marshal(struct {
		Keys []telemetryTestPrivateJWK `json:"keys"`
	}{Keys: keys})
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
		OrgID:         "org-test",
		AppID:         "app-test",
		InstallID:     "install-test",
		RunnerID:      "runner-test",
		RelayEndpoint: "https://relay.example.com/acme",
	}

	raw, err := issuer.issue(principal, true)
	require.NoError(t, err)

	claims := &telemetryAccessTokenClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return &issuer.privateKey.PublicKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithIssuer(issuer.issuer), jwt.WithAudience("https://relay.example.com/acme"), jwt.WithTimeFunc(func() time.Time { return now }))
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
	require.Equal(t, []any{"https://relay.example.com/acme"}, wireClaims["aud"])
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

func TestRunnerServiceTelemetryTokenIssuerConfiguration(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		svc, err := New(Params{})

		require.NoError(t, err)
		require.Nil(t, svc.telemetryTokenIssuer)
	})

	t.Run("invalid", func(t *testing.T) {
		svc, err := New(Params{Cfg: &internal.Config{
			PublicAPIURL:  "https://ctl.example.com",
			TelemetryJWKS: "not-json",
		}})

		require.Nil(t, svc)
		require.ErrorContains(t, err, "initialize telemetry token issuer")
		require.ErrorContains(t, err, "decode JWKS")
	})
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
