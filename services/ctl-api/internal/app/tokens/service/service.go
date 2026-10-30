package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type Params struct {
	fx.In

	DB         *gorm.DB `name:"psql"`
	AcctClient *account.Client
	Helpers    *accountshelpers.Helpers
}

type service struct {
	db         *gorm.DB
	acctClient *account.Client
	helpers    *accountshelpers.Helpers
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db:         params.DB,
		acctClient: params.AcctClient,
		helpers:    params.Helpers,
	}
}

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	tokens := ge.Group("/v1/tokens")
	{
		require.NotFederatedTokenRoute(tokens, "create API tokens").POST("", s.CreateToken)
		tokens.GET("", s.ListTokens)
		tokens.DELETE("/:token_id", s.DeleteToken)
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
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
