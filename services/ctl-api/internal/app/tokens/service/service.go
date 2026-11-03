package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type Params struct {
	fx.In

	DB         *gorm.DB `name:"psql"`
	Logger     *zap.Logger
	AcctClient *account.Client
	Helpers    *accountshelpers.Helpers
}

type service struct {
	db         *gorm.DB
	l          *zap.Logger
	acctClient *account.Client
	helpers    *accountshelpers.Helpers
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db:         params.DB,
		l:          params.Logger,
		acctClient: params.AcctClient,
		helpers:    params.Helpers,
	}
}

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	tokens := require.CustomerRoute(ge.Group("/v1/tokens"))
	{
		contract := require.MutationContract{Permission: permissions.OrgWrite, Target: require.ReadTarget{Name: "org_id", Kind: permissions.KindOrg}}
		require.MutationRoute(require.NotFederatedTokenRoute(tokens, "create API tokens"), s.db, contract).HandlePrepared("POST", "", s.CreateToken)
		require.ReadRoute(tokens, s.db, require.ReadContract{
			Permission: permissions.OrgWrite, Targets: []require.ReadTarget{{Name: "org_id", Kind: permissions.KindOrg}},
		}).GET("", s.ListTokens)
		require.MutationRoute(tokens, s.db, contract).Handle("DELETE", "/:token_id", s.DeleteToken)
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
