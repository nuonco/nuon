package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// @ID GetTelemetryJWKS
// @Summary Get telemetry JWT public keys
// @Description Returns the public RSA keys used to verify BYOC telemetry access tokens.
// @Tags runners
// @Produce json
// @Success 200 {object} TelemetryJSONWebKeySet
// @Failure 503 {object} stderr.ErrResponse
// @Router /.well-known/jwks.json [GET]
func (s *service) GetTelemetryJWKS(ctx *gin.Context) {
	if s.telemetryTokenIssuer == nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error:       "telemetry public keys are unavailable",
			Description: "telemetry public keys are unavailable",
		})
		return
	}

	ctx.Header("Cache-Control", "public, max-age=300")
	ctx.JSON(http.StatusOK, s.telemetryTokenIssuer.publicJWKS())
}
