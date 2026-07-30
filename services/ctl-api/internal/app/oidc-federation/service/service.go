package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

type Params struct {
	fx.In

	V           *validator.Validate
	DB          *gorm.DB `name:"psql"`
	L           *zap.Logger
	Cfg         *internal.Config
	AuthzClient *authz.Client
	MW          metrics.Writer
}

type service struct {
	v           *validator.Validate
	l           *zap.Logger
	db          *gorm.DB
	cfg         *internal.Config
	authzClient *authz.Client
	mw          metrics.Writer
	jwks        *jwksProviderCache
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		v:           params.V,
		l:           params.L,
		db:          params.DB,
		cfg:         params.Cfg,
		authzClient: params.AuthzClient,
		mw:          params.MW,
		jwks:        newJWKSProviderCache(),
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	oidc := api.Group("/v1/oidc")
	{
		// unauthenticated token exchange; listed in the public endpoint list
		oidc.POST("/token", s.ExchangeOIDCToken)

		policies := oidc.Group("/trust-policies")
		{
			policies.POST("", s.CreateOIDCTrustPolicy)
			policies.GET("", s.ListOIDCTrustPolicies)
			policies.GET("/:policy_id", s.GetOIDCTrustPolicy)
			policies.PATCH("/:policy_id", s.UpdateOIDCTrustPolicy)
			policies.DELETE("/:policy_id", s.DeleteOIDCTrustPolicy)
		}
	}

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
