package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type Params struct {
	fx.In

	Cfg *internal.Config
}

type service struct {
	issuer     *oidcissuer.Issuer
	publicKeys oidcissuer.JWKS
}

var _ api.Service = (*service)(nil)

func New(params Params) (*service, error) {
	s := &service{}
	if params.Cfg == nil || params.Cfg.TelemetryJWKS == "" {
		return s, nil
	}
	privateKey, keyID, publicKeys, err := oidcissuer.ParseJWKS(params.Cfg.TelemetryJWKS)
	if err != nil {
		return nil, fmt.Errorf("initialize OIDC public keys: %w", err)
	}
	issuer, err := oidcissuer.New(params.Cfg.PublicAPIURL, privateKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("initialize OIDC issuer: %w", err)
	}
	s.issuer = issuer
	s.publicKeys = publicKeys
	return s, nil
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	api.GET(oidcissuer.JWKSPath, s.GetJWKS)
	api.GET("/.well-known/openid-configuration", s.GetOpenIDConfiguration)
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
