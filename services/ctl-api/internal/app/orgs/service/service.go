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
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
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
	AuthzClient     *authz.Client
	RunnersHelpers  *runnershelpers.Helpers
	AcctClient      *account.Client
	AnalyticsClient analytics.Writer
	Helpers         *helpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	VCSHelpers      *vcshelpers.Helpers
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
	runnersHelpers  *runnershelpers.Helpers
	acctClient      *account.Client
	analyticsClient analytics.Writer
	helpers         *helpers.Helpers
	accountsHelpers *accountshelpers.Helpers
	vcsHelpers      *vcshelpers.Helpers
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
		require.NotFederatedTokenRoute(orgs, "create organizations").POST("", s.CreateOrg)
		orgs.GET("", s.GetCurrentUserOrgs)
		s.orgRoute(orgs, permissions.OrgFeaturesRead).GET("/features", s.GetOrgFeatures)

		// update your current org
		current := orgs.Group("/current")
		currentAccess := require.Routes(current, permissions.APICtl)
		{
			currentAccess.GET("", require.Point(permissions.OrgInspect, s.orgTarget), s.GetOrg)
			s.orgRoute(current, permissions.OrgDelete).DELETE("", s.DeleteOrg)
			s.orgRoute(current, permissions.OrgUpdate).PATCH("", s.UpdateOrg)
			s.orgRoute(current, permissions.OrgTelemetryUpdate).PATCH("/telemetry", s.UpdateOrgTelemetry)
			delegate := require.UnboundedCredentialRoute(current)
			s.orgRoute(delegate, permissions.OrgMemberAdd).POST("/user", s.CreateUser)
			s.orgRoute(current, permissions.OrgMemberRemove).POST("/remove-user", s.RemoveUser)

			// accounts
			s.orgRoute(current, permissions.OrgMembersRead).GET("/accounts", s.GetOrgAccounts)
			s.orgRoute(current, permissions.OrgMembersRead).GET("/members", s.GetOrgMembers)
			s.orgRoute(delegate, permissions.OrgMemberRoleUpdate).PATCH("/accounts/:account_id/role", s.UpdateOrgAccountRole)

			// invites
			invites := current.Group("/invites")
			{
				s.orgRoute(invites, permissions.OrgInvitesRead).GET("", s.GetOrgInvites)
				issue := require.UnboundedCredentialRoute(invites)
				s.orgRoute(issue, permissions.OrgInviteCreate).POST("", s.CreateOrgInvite)
				s.orgRoute(issue, permissions.OrgInviteCreate).POST("/:invite_id/resend", s.ResendOrgInvite)
				s.orgRoute(invites, permissions.OrgInviteRevoke).POST("/:invite_id/revoke", s.RevokeOrgInvite)
			}

			s.orgRoute(current, permissions.OrgStatsRead).GET("/stats", s.GetOrgStats)

			// features
			s.orgRoute(current, permissions.OrgFeaturesRead).GET("/features", s.GetCurrentOrgFeatures)
			s.orgRoute(current, permissions.OrgFeaturesUpdate).PATCH("/features", s.UpdateOrgFeatures) // requires user-managed-features flag

			// webhooks
			s.orgRoute(current, permissions.OrgWebhooksRead).GET("/webhooks", s.GetCurrentOrgWebhooks)
			s.orgRoute(current, permissions.OrgWebhookCreate).POST("/webhooks", s.CreateCurrentOrgWebhook)
			s.orgRoute(current, permissions.OrgWebhookUpdate).PATCH("/webhooks/:webhook_id", s.UpdateCurrentOrgWebhook)
			s.orgRoute(current, permissions.OrgWebhookDelete).DELETE("/webhooks/:webhook_id", s.DeleteCurrentOrgWebhook)
		}
	}

	return nil
}

func (s *service) orgRoute(router gin.IRouter, operation permissions.PermissionID) *gin.RouterGroup {
	routerAccess := require.Routes(router, permissions.APICtl)
	return routerAccess.For(require.Point(operation, s.orgTarget))
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
		orgs.POST("/admin-reconcile-roles", s.AdminReconcileOrgRoles)

		// org features (all orgs)
		orgs.GET("/admin-features", s.AdminGetOrgFeatures)
		orgs.PATCH("/admin-features", s.AdminUpdateOrgsFeatures)
		orgs.PATCH("/admin-toggle-feature", s.AdminToggleOrgsFeature)

		// org-specific admin routes
		org := orgs.Group("/:org_id")
		{
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
			org.POST("/admin-enable-email-notifications", s.AdminEnableEmailNotifications)
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
			org.POST("/migrate-crons-namespaces", s.MigrateCronsNamespaces)
			org.POST("/migrate-healthcheck-sweeps", s.MigrateHealthcheckSweeps)
			org.POST("/admin-restart-runners", s.AdminRestartRunners)
			org.POST("/admin-graceful-shutdown-processes", s.AdminGracefulShutdownOrgRunnerProcesses)
			org.POST("/admin-force-shutdown-processes", s.AdminForceShutdownOrgRunnerProcesses)
			org.POST("/admin-labels", s.AdminAddOrgLabels)
			org.DELETE("/admin-labels", s.AdminRemoveOrgLabels)
			org.PATCH("/admin-features", s.AdminUpdateOrgFeatures)
			org.POST("/admin-backfill-phone-home", s.AdminBackfillOrgPhoneHome)
			org.POST("/admin-add-tags", s.AdminAddTags)
			org.POST("/admin-remove-tags", s.AdminRemoveTags)
		}
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	require.Routes(require.OrgRoute(api), permissions.APIRunner).
		GET("/v1/orgs/current", require.Organization(permissions.OrgInspect), s.GetOrg)
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
		authzClient:     params.AuthzClient,
		runnersHelpers:  params.RunnersHelpers,
		analyticsClient: params.AnalyticsClient,
		acctClient:      params.AcctClient,
		helpers:         params.Helpers,
		accountsHelpers: params.AccountsHelpers,
		vcsHelpers:      params.VCSHelpers,
		features:        params.Features,
		queueClient:     params.QueueClient,
		emitterClient:   params.EmitterClient,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
