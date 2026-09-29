package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// @ID GetOpenIDConfiguration
// @Summary Get OIDC discovery document
// @Description Returns the OIDC discovery document cloud providers use to federate to this control plane.
// @Tags oidc
// @Produce json
// @Success 200 {object} oidcissuer.DiscoveryDocument
// @Failure 503 {object} stderr.ErrResponse
// @Router /.well-known/openid-configuration [GET]
func (s *service) GetOpenIDConfiguration(ctx *gin.Context) {
	if s.issuer == nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error:       "OIDC discovery is unavailable",
			Description: "OIDC discovery is unavailable",
		})
		return
	}

	ctx.JSON(http.StatusOK, s.issuer.DiscoveryDocument())
}
