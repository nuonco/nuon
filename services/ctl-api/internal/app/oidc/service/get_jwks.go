package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

// Aliases keep the swagger model names stable now that JWKS parsing and
// validation live in the shared oidcissuer package.
type TelemetryJSONWebKey = oidcissuer.JWK
type TelemetryJSONWebKeySet = oidcissuer.JWKS

// @ID GetTelemetryJWKS
// @Summary Get OIDC signing public keys
// @Description Returns the public RSA keys used to verify cloud federation and telemetry tokens.
// @Tags oidc
// @Produce json
// @Success 200 {object} TelemetryJSONWebKeySet
// @Failure 503 {object} stderr.ErrResponse
// @Router /.well-known/jwks.json [GET]
func (s *service) GetJWKS(ctx *gin.Context) {
	if s.issuer == nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error:       "telemetry public keys are unavailable",
			Description: "telemetry public keys are unavailable",
		})
		return
	}

	ctx.Header("Cache-Control", "public, max-age=300")
	ctx.JSON(http.StatusOK, s.publicKeys)
}
