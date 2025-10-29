package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	accountshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/accounts/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"
)

type Params struct {
	fx.In

	V               *validator.Validate
	DB              *gorm.DB `name:"psql"`
	MW              metrics.Writer
	L               *zap.Logger
	Cfg             *internal.Config
	EvClient        eventloop.Client
	AuthzClient     *authz.Client
	RunnersHelpers  *runnershelpers.Helpers
	AcctClient      *account.Client
	AnalyticsClient analytics.Writer
	Helpers         *helpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	Features        *features.Features
	EndpointAudit   *api.EndpointAudit
}

type service struct {
	api.RouteRegister
	v               *validator.Validate
	l               *zap.Logger
	db              *gorm.DB
	mw              metrics.Writer
	cfg             *internal.Config
	authzClient     *authz.Client
	evClient        eventloop.Client
	runnersHelpers  *runnershelpers.Helpers
	acctClient      *account.Client
	analyticsClient analytics.Writer
	helpers         *helpers.Helpers
	accountsHelpers *accountshelpers.Helpers
	features        *features.Features
	endpointAudit   *api.EndpointAudit
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	// global routes
	orgs := ge.Group("/v1/orgs")
	{
		orgs.POST("", s.CreateOrg)
		orgs.GET("", s.GetCurrentUserOrgs)

		// update your current org
		current := orgs.Group("/current")
		{
			current.GET("", s.GetOrg)
			current.DELETE("", s.DeleteOrg)
			current.PATCH("", s.UpdateOrg)
			current.POST("/user", s.CreateUser)
			current.POST("/remove-user", s.RemoveUser)

			// accounts
			current.GET("/accounts", s.GetOrgAccounts)

			// invites
			invites := current.Group("/invites")
			{
				invites.GET("", s.GetOrgInvites)
				invites.POST("", s.CreateOrgInvite)
			}

			// runners
			current.GET("/runner-group", s.GetOrgRunnerGroup)
		}
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	orgs := api.Group("/v1/orgs")
	{
		// global org operations
		orgs.GET("", s.GetAllOrgs)
		orgs.GET("/admin-get", s.AdminGetOrg)
		orgs.POST("/admin-delete-canarys", s.AdminDeleteCanaryOrgs)
		orgs.POST("/admin-delete-integrations", s.AdminDeleteIntegrationOrgs)
		orgs.POST("/admin-restart-all", s.RestartAllOrgs)

		// org features (all orgs)
		orgs.GET("/admin-features", s.AdminGetOrgFeatures)
		orgs.PATCH("/admin-features", s.AdminUpdateOrgsFeatures)

		// org-specific admin routes
		org := orgs.Group("/:org_id")
		{
			org.GET("/admin-get-runner", s.AdminGetOrgRunner)
			org.POST("/admin-add-user", s.CreateOrgUser)
			org.POST("/admin-support-users", s.CreateSupportUsers)
			org.POST("/admin-remove-support-users", s.RemoveSupportUsers)
			org.POST("/admin-delete", s.AdminDeleteOrg)
			org.POST("/admin-reprovision", s.AdminReprovisionOrg)
			org.POST("/admin-deprovision", s.AdminDeprovisionOrg)
			org.POST("/admin-restart", s.RestartOrg)
			org.POST("/admin-restart-children", s.RestartOrgChildren)
			org.POST("/admin-rename", s.AdminRenameOrg)
			org.POST("/admin-internal-slack-webhook-url", s.AdminSetInternalSlackWebhookURLOrg)
			org.POST("/admin-customer-slack-webhook-url", s.AdminSetCustomerSlackWebhookURLOrg)
			org.POST("/admin-add-vcs-connection", s.AdminAddVCSConnection)
			org.POST("/admin-service-account", s.AdminCreateServiceAccount)
			org.POST("/admin-add-logo", s.AdminAddLogo)
			org.POST("/admin-migrate", s.AdminMigrateOrg)
			org.POST("/admin-debug-mode", s.AdminDebugModeOrg)
			org.POST("/admin-add-priority", s.AdminAddPriority)
			org.POST("/admin-forget", s.AdminForgetOrg)
			org.POST("/admin-force-sandbox-mode", s.AdminForceSandboxMode)
			org.POST("/admin-restart-runners", s.AdminRestartRunners)
			org.PATCH("/admin-features", s.AdminUpdateOrgFeatures)
		}
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs/current", s.GetOrg)
	return nil
}

func New(params Params) *service {
	return &service{
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		l:               params.L,
		v:               params.V,
		db:              params.DB,
		mw:              params.MW,
		cfg:             params.Cfg,
		evClient:        params.EvClient,
		authzClient:     params.AuthzClient,
		runnersHelpers:  params.RunnersHelpers,
		analyticsClient: params.AnalyticsClient,
		acctClient:      params.AcctClient,
		helpers:         params.Helpers,
		accountsHelpers: params.AccountsHelpers,
		features:        params.Features,
	}
}
