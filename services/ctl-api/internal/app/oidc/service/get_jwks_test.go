package service

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func TestGetJWKS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	encode := func(value *big.Int) string { return base64.RawURLEncoding.EncodeToString(value.Bytes()) }
	configured, err := json.Marshal(map[string]any{"keys": []map[string]string{{
		"kty": "RSA", "kid": "telemetry-key-1", "use": "sig", "alg": "RS256",
		"n": encode(key.N), "e": encode(big.NewInt(int64(key.E))),
		"d": encode(key.D), "p": encode(key.Primes[0]), "q": encode(key.Primes[1]),
	}}})
	require.NoError(t, err)
	svc, err := New(Params{Cfg: &internal.Config{PublicAPIURL: "https://ctl.example.com/", TelemetryJWKS: string(configured)}})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)

	svc.GetJWKS(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "public, max-age=300", recorder.Header().Get("Cache-Control"))
	var response oidcissuer.JWKS
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Keys, 1)
	require.Equal(t, oidcissuer.JWK{
		KeyType: "RSA", KeyID: "telemetry-key-1", Use: "sig", Algorithm: "RS256",
		Modulus: encode(key.N), Exponent: encode(big.NewInt(int64(key.E))),
	}, response.Keys[0])
	var published map[string][]map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &published))
	require.NotContains(t, published["keys"][0], "d")
	require.NotContains(t, published["keys"][0], "p")
	require.NotContains(t, published["keys"][0], "q")
}

func TestOIDCEndpointsUnavailableWithoutIssuer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &service{}
	for _, test := range []struct {
		name   string
		path   string
		invoke func(*gin.Context)
	}{
		{name: "public keys", path: "/.well-known/jwks.json", invoke: svc.GetJWKS},
		{name: "openid configuration", path: "/.well-known/openid-configuration", invoke: svc.GetOpenIDConfiguration},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, test.path, nil)
			test.invoke(ctx)
			require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
		})
	}
}
