package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetrytoken"
)

type CreateInstallTelemetryAccessTokenResponse = telemetrytoken.AccessTokenResponse

// @ID CreateInstallTelemetryAccessToken
// @Summary Create a relay access token for an install
// @Description Requires create permission on the install's telemetry resource. Returns a ten-minute telemetry:write JWT identifying the authenticated account and install, bound to the current relay endpoint. Disabled telemetry refuses issuance; existing JWTs remain valid until expiry.
// @Tags installs/runner
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Param relay_endpoint query string true "Exact relay endpoint from telemetry settings"
// @Success 200 {object} CreateInstallTelemetryAccessTokenResponse
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Failure 503 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/telemetry/access-token [post]
func (s *service) CreateInstallTelemetryAccessToken(ctx *gin.Context) {
	if s.telemetryTokenIssuer == nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error: "telemetry token issuance is unavailable", Description: "telemetry token issuance is unavailable",
		})
		return
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	requestedEndpoint := ctx.Query("relay_endpoint")
	if requestedEndpoint == "" {
		ctx.Error(stderr.ErrUser{
			Err: errors.New("relay_endpoint is required"), Description: "fetch telemetry settings and supply relay_endpoint",
		})
		return
	}
	principal, err := s.helpers.GetInstallTelemetryTokenPrincipal(ctx, orgID, ctx.Param("install_id"), acct.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := app.ValidateTelemetryRelayEndpoint(principal.RelayEndpoint); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error: "telemetry relay is unavailable", Description: "telemetry relay is unavailable",
		})
		return
	}
	if requestedEndpoint != principal.RelayEndpoint {
		ctx.JSON(http.StatusConflict, stderr.ErrResponse{
			Error: "telemetry relay endpoint does not match current settings", Description: "refresh telemetry settings before requesting a telemetry token",
		})
		return
	}
	accessToken, err := s.telemetryTokenIssuer.Issue(principal)
	if err != nil {
		ctx.Error(fmt.Errorf("create install telemetry access token: %w", err))
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Pragma", "no-cache")
	ctx.JSON(http.StatusOK, CreateInstallTelemetryAccessTokenResponse{
		AccessToken: accessToken, TokenType: "Bearer", ExpiresIn: int64(telemetrytoken.Lifetime.Seconds()),
	})
}
