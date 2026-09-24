package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestNewTelemetryRelayEndpoint(t *testing.T) {
	issuer := &telemetryTokenIssuer{}

	t.Run("disabled", func(t *testing.T) {
		endpoint, err := newTelemetryRelayEndpoint(&internal.Config{}, nil)
		require.NoError(t, err)
		require.Empty(t, endpoint)
	})

	t.Run("configured", func(t *testing.T) {
		endpoint, err := newTelemetryRelayEndpoint(&internal.Config{TelemetryRelayEndpoint: "https://telemetry.example.com/otlp"}, issuer)
		require.NoError(t, err)
		require.Equal(t, "https://telemetry.example.com/otlp", endpoint)
	})

	t.Run("requires issuer", func(t *testing.T) {
		_, err := newTelemetryRelayEndpoint(&internal.Config{TelemetryRelayEndpoint: "https://telemetry.example.com"}, nil)
		require.ErrorContains(t, err, "requires a configured telemetry token issuer")
	})

	for name, endpoint := range map[string]string{
		"http":                  "http://telemetry.example.com",
		"userinfo":              "https://user@telemetry.example.com",
		"query":                 "https://telemetry.example.com?token=value",
		"fragment":              "https://telemetry.example.com#fragment",
		"environment expansion": "https://${TELEMETRY_HOST}",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := newTelemetryRelayEndpoint(&internal.Config{TelemetryRelayEndpoint: endpoint}, issuer)
			require.Error(t, err)
		})
	}
}

func TestNewRejectsInvalidTelemetryRelayConfiguration(t *testing.T) {
	key := telemetryTestRSAKey(t)
	_, err := New(Params{
		Cfg: &internal.Config{
			PublicAPIURL:           "https://ctl.example.com",
			TelemetryJWKS:          telemetryTestJWKS(t, telemetryTestJWK(key, "telemetry-key-1", true)),
			TelemetryRelayEndpoint: "http://telemetry.example.com",
		},
		L: zap.NewNop(),
	})
	require.ErrorContains(t, err, "invalid telemetry relay configuration")
}

func TestCreateTelemetryAccessTokenUnavailableWithoutRelay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	issuer, _ := newTelemetryTestTokenIssuer(t)
	svc := &service{telemetryTokenIssuer: issuer}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/telemetry/access-token", nil)

	svc.CreateTelemetryAccessToken(ctx)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}
