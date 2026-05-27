package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/analytics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
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
	QueueClient     *queueclient.Client
	EmitterClient   *emitterclient.Client
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
	queueClient     *queueclient.Client
	emitterClient   *emitterclient.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	// global routes
	orgs := ge.Group("/v1/orgs")
	{
		orgs.POST("", s.CreateOrg)
		orgs.GET("", s.GetCurrentUserOrgs)
		orgs.GET("/features", s.GetOrgFeatures)

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
				invites.POST("/:invite_id/resend", s.ResendOrgInvite)
				invites.POST("/:invite_id/revoke", s.RevokeOrgInvite)
			}

			// runners
			current.GET("/runner-group", s.GetOrgRunnerGroup)

			current.GET("/stats", s.GetOrgStats)

			// features
			current.GET("/features", s.GetCurrentOrgFeatures)
			current.PATCH("/features", s.UpdateOrgFeatures) // requires user-managed-features flag

			// webhooks
			current.GET("/webhooks", s.GetCurrentOrgWebhooks)
			current.POST("/webhooks", s.CreateCurrentOrgWebhook)
			current.PATCH("/webhooks/:webhook_id", s.UpdateCurrentOrgWebhook)
			current.DELETE("/webhooks/:webhook_id", s.DeleteCurrentOrgWebhook)
		}
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	orgs := api.Group("/v1/orgs")
	{
		// global org operations
		orgs.GET("", s.GetAllOrgs)
		orgs.GET("/details", s.AdminListOrgsDetails)
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
			org.POST("/admin-restart-queues", s.RestartOrgQueues)
			org.POST("/admin-force-restart-queues", s.ForceRestartOrgQueues)
			org.POST("/admin-migrate-queues", s.AdminMigrateOrgQueues)
			org.POST("/admin-restart-runners", s.AdminRestartRunners)
			org.POST("/admin-graceful-shutdown-processes", s.AdminGracefulShutdownOrgRunnerProcesses)
			org.POST("/admin-force-shutdown-processes", s.AdminForceShutdownOrgRunnerProcesses)
			org.POST("/admin-labels", s.AdminAddOrgLabels)
			org.DELETE("/admin-labels", s.AdminRemoveOrgLabels)
			org.PATCH("/admin-features", s.AdminUpdateOrgFeatures)
			org.POST("/admin-add-tags", s.AdminAddTags)
			org.POST("/admin-remove-tags", s.AdminRemoveTags)
		}
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs/current", s.GetOrg)
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
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
		queueClient:     params.QueueClient,
		emitterClient:   params.EmitterClient,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
