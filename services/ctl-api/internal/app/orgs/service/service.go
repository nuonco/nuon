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
	ge.POST("/v1/orgs", s.CreateOrg)
	ge.GET("/v1/orgs", s.GetCurrentUserOrgs)

	// update your current org
	ge.GET("/v1/orgs/current", s.GetOrg)
	ge.DELETE("/v1/orgs/current", s.DeleteOrg)
	ge.PATCH("/v1/orgs/current", s.UpdateOrg)
	ge.POST("/v1/orgs/current/user", s.CreateUser)
	ge.POST("/v1/orgs/current/remove-user", s.RemoveUser)

	// accounts
	ge.GET("/v1/orgs/current/accounts", s.GetOrgAccounts)

	// invites
	ge.GET("/v1/orgs/current/invites", s.GetOrgInvites)
	ge.POST("/v1/orgs/current/invites", s.CreateOrgInvite)

	// runners
	ge.GET("/v1/orgs/current/runner-group", s.GetOrgRunnerGroup)

	// user journeys
	ge.GET("/v1/orgs/current/user-journeys", s.GetOrgUserJourneys)
	ge.POST("/v1/orgs/current/user-journeys", s.CreateUserJourney)
	ge.PATCH("/v1/orgs/current/user-journeys/:journey_name/steps/:step_name", s.UpdateUserJourneyStep)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs", s.GetAllOrgs)
	api.GET("/v1/orgs/admin-get", s.AdminGetOrg)
	api.GET("/v1/orgs/:org_id/admin-get-runner", s.AdminGetOrgRunner)
	api.POST("/v1/orgs/admin-delete-canarys", s.AdminDeleteCanaryOrgs)
	api.POST("/v1/orgs/admin-delete-integrations", s.AdminDeleteIntegrationOrgs)

	api.POST("/v1/orgs/:org_id/admin-add-user", s.CreateOrgUser)
	api.POST("/v1/orgs/:org_id/admin-support-users", s.CreateSupportUsers)
	api.POST("/v1/orgs/:org_id/admin-remove-support-users", s.RemoveSupportUsers)
	api.POST("/v1/orgs/:org_id/admin-delete", s.AdminDeleteOrg)
	api.POST("/v1/orgs/:org_id/admin-reprovision", s.AdminReprovisionOrg)
	api.POST("/v1/orgs/:org_id/admin-deprovision", s.AdminDeprovisionOrg)
	api.POST("/v1/orgs/:org_id/admin-restart", s.RestartOrg)
	api.POST("/v1/orgs/:org_id/admin-restart-children", s.RestartOrgChildren)
	api.POST("/v1/orgs/:org_id/admin-rename", s.AdminRenameOrg)
	api.POST("/v1/orgs/:org_id/admin-internal-slack-webhook-url", s.AdminSetInternalSlackWebhookURLOrg)
	api.POST("/v1/orgs/:org_id/admin-customer-slack-webhook-url", s.AdminSetCustomerSlackWebhookURLOrg)
	api.POST("/v1/orgs/:org_id/admin-add-vcs-connection", s.AdminAddVCSConnection)
	api.POST("/v1/orgs/:org_id/admin-service-account", s.AdminCreateServiceAccount)
	api.POST("/v1/orgs/:org_id/admin-add-logo", s.AdminAddLogo)
	api.POST("/v1/orgs/:org_id/admin-migrate", s.AdminMigrateOrg)
	api.POST("/v1/orgs/:org_id/admin-debug-mode", s.AdminDebugModeOrg)
	api.POST("/v1/orgs/:org_id/admin-add-priority", s.AdminAddPriority)
	api.POST("/v1/orgs/:org_id/admin-forget", s.AdminForgetOrg)
	api.POST("/v1/orgs/:org_id/admin-force-sandbox-mode", s.AdminForceSandboxMode)
	api.POST("/v1/orgs/:org_id/admin-restart-runners", s.AdminRestartRunners)
	api.PATCH("/v1/orgs/:org_id/admin-features", s.AdminUpdateOrgFeatures)

	// for updating all
	api.GET("/v1/orgs/admin-features", s.AdminGetOrgFeatures)
	api.PATCH("/v1/orgs/admin-features", s.AdminUpdateOrgsFeatures)
	api.POST("/v1/orgs/admin-restart-all", s.RestartAllOrgs)

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
