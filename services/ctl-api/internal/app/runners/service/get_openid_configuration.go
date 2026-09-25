package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

// newOIDCIssuer builds the issuer cloud federation trusts, deriving the issuer
// from the public API URL so it matches the discovery document exactly. It
// shares the telemetry signing key; the audience keeps these tokens distinct
// from telemetry access tokens.
func newOIDCIssuer(cfg *internal.Config, issuer *telemetryTokenIssuer) (*oidcissuer.Issuer, error) {
	if issuer == nil {
		return nil, nil
	}

	oidcIssuer, err := oidcissuer.New(cfg.PublicAPIURL, issuer.privateKey, issuer.keyID)
	if err != nil {
		return nil, fmt.Errorf("initialize OIDC issuer: %w", err)
	}
	return oidcIssuer, nil
}

// @ID GetOpenIDConfiguration
// @Summary Get OIDC discovery document
// @Description Returns the OIDC discovery document cloud providers use to federate to this control plane.
// @Tags runners
// @Produce json
// @Success 200 {object} oidcissuer.DiscoveryDocument
// @Failure 503 {object} stderr.ErrResponse
// @Router /.well-known/openid-configuration [GET]
func (s *service) GetOpenIDConfiguration(ctx *gin.Context) {
	if s.oidcIssuer == nil {
		ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
			Error:       "OIDC discovery is unavailable",
			Description: "OIDC discovery is unavailable",
		})
		return
	}

	ctx.JSON(http.StatusOK, s.oidcIssuer.DiscoveryDocument())
}
