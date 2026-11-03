package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/productanalytics"
)

type Params struct {
	fx.In

	DB               *gorm.DB `name:"psql"`
	Logger           *zap.Logger
	AcctClient       *account.Client
	AuthzClient      *authz.Client
	ProductAnalytics *productanalytics.Client
	EndpointAudit    *api.EndpointAudit
	Helpers          *accountshelpers.Helpers
}

type service struct {
	api.RouteRegister
	db               *gorm.DB
	l                *zap.Logger
	acctClient       *account.Client
	authzClient      *authz.Client
	productAnalytics *productanalytics.Client
	helpers          *accountshelpers.Helpers
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		RouteRegister:    api.RouteRegister{EndpointAudit: params.EndpointAudit},
		db:               params.DB,
		l:                params.Logger,
		acctClient:       params.AcctClient,
		authzClient:      params.AuthzClient,
		productAnalytics: params.ProductAnalytics,
		helpers:          params.Helpers,
	}
}
func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	// accounts
	account := ge.Group("/v1/account")
	{
		account.GET("", s.GetCurrentAccount)

		s.POST(require.NotFederatedTokenRoute(require.CustomerRoute(account), "create API tokens"), "/static-token", s.CreateStaticToken, api.APIContextTypePublic, true)
		s.GET(require.CustomerRoute(account), "/static-tokens", s.ListStaticTokens, api.APIContextTypePublic, true)
		s.DELETE(require.CustomerRoute(account), "/static-tokens/:token_id", s.DeleteStaticToken, api.APIContextTypePublic, true)

		// user journeys
		userJourneys := account.Group("/user-journeys")
		{
			userJourneys.GET("", s.GetUserJourneys)
			userJourneys.POST("", s.CreateUserJourney)
			userJourneys.PATCH("/:journey_name/steps/:step_name", s.UpdateUserJourneyStep)
			userJourneys.POST("/:journey_name/reset", s.ResetUserJourney)
			userJourneys.POST("/:journey_name/complete", s.CompleteUserJourney)
		}
	}

	// roles
	require.CustomerReadRoute(ge, s.db, s.l, permissions.OrgWrite, permissions.KindOrg, "").GET("/v1/roles", s.ListRoles)

	// service accounts
	serviceAccounts := require.CustomerRoute(ge.Group("/v1/service-accounts"))
	{
		serviceAccounts.GET("", s.ListServiceAccounts)
		serviceAccounts.GET("/purposes", s.ListServiceAccountPurposes)
		require.MutationRoute(serviceAccounts, s.db, require.MutationContract{
			Permission: permissions.OrgWrite, Target: require.ReadTarget{Name: "org_id", Kind: permissions.KindOrg},
		}).Handle("POST", "", s.CreateServiceAccount)
		mutations := require.MutationRoute(serviceAccounts, s.db, require.MutationContract{
			Permission: permissions.OrgWrite, Target: require.ReadTarget{Name: "org_id", Kind: permissions.KindOrg},
			SubjectAccountParam: "account_id",
		})
		mutations.Handle("PATCH", "/:account_id", s.UpdateServiceAccount)
		mutations.Handle("PATCH", "/:account_id/role", s.UpdateServiceAccountRole)
		mutations.Handle("DELETE", "/:account_id", s.DeleteServiceAccount)
		s.POST(require.NotFederatedTokenRoute(serviceAccounts, "create API tokens"), "/:account_id/tokens", s.CreateServiceAccountToken, api.APIContextTypePublic, true)
	}

	// auth/me - registered here instead of authservice so it's available in PublicServicesModule
	auth := ge.Group("/v1/auth")
	{
		auth.GET("/me", s.GetAuthMe)
		auth.GET("/validate", s.ValidateToken)
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// No internal routes for accounts service at this time
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	// No runner routes for accounts service at this time
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) getAccount(ctx *gin.Context, accountID string) (*app.Account, error) {
	var account app.Account

	res := s.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Policies").
		Preload("Roles.Org").
		Where("id = ?", accountID).
		First(&account)

	if res.Error != nil {
		return nil, res.Error
	}

	return &account, nil
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
