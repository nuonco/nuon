package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateTelemetryAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// @ID CreateTelemetryAccessToken
// @Summary Create a telemetry access token
// @Description Creates a short-lived, install-runner-scoped JWT for the resolved telemetry relay. Org overrides require relay_endpoint to match current settings.
// @Param relay_endpoint query string false "Relay endpoint from runner settings; prevents issuing a token for a stale destination"
// @Tags runners/runner
// @Produce json
// @Security APIKey
// @Security OrgID
// @Success 200 {object} CreateTelemetryAccessTokenResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Failure 503 {object} stderr.ErrResponse
// @Router /v1/telemetry/access-token [POST]
func (s *service) CreateTelemetryAccessToken(ctx *gin.Context) {
	requestedEndpoint := ctx.Query("relay_endpoint")
	if s.telemetryTokenIssuer == nil || (s.telemetryRelayEndpoint == "" && requestedEndpoint == "") {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error:       "telemetry token issuance is unavailable",
			Description: "telemetry token issuance is unavailable",
		})
		return
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         fmt.Errorf("get telemetry runner account: %w", err),
			Description: "unable to create telemetry access token",
		})
		return
	}
	principal, err := s.resolveTelemetryRunnerPrincipal(ctx, acct)
	if err != nil {
		ctx.Error(err)
		return
	}

	if (requestedEndpoint != "" && requestedEndpoint != principal.RelayEndpoint) || (requestedEndpoint == "" && principal.RelayEndpoint != s.telemetryRelayEndpoint) {
		ctx.JSON(http.StatusConflict, stderr.ErrResponse{
			Error:       "telemetry relay endpoint does not match current settings",
			Description: "refresh runner settings before requesting a telemetry token",
		})
		return
	}

	accessToken, err := s.telemetryTokenIssuer.issue(principal)
	if err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         fmt.Errorf("create telemetry access token: %w", err),
			Description: "unable to create telemetry access token",
		})
		return
	}

	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Pragma", "no-cache")
	ctx.JSON(http.StatusOK, CreateTelemetryAccessTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(telemetryTokenLifetime.Seconds()),
	})
}
