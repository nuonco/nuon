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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
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
}

type service struct {
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
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// global routes
	api.POST("/v1/orgs", s.CreateOrg)
	api.GET("/v1/orgs", s.GetCurrentUserOrgs)

	// update your current org
	api.GET("/v1/orgs/current", s.GetOrg)
	api.DELETE("/v1/orgs/current", s.DeleteOrg)
	api.PATCH("/v1/orgs/current", s.UpdateOrg)
	api.POST("/v1/orgs/current/user", s.CreateUser)

	// invites
	api.GET("/v1/orgs/current/invites", s.GetOrgInvites)
	api.POST("/v1/orgs/current/invites", s.CreateOrgInvite)

	// runners
	api.GET("/v1/orgs/current/runner-group", s.GetOrgRunnerGroup)

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

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs/current", s.GetOrg)
	return nil
}

func New(params Params) *service {
	return &service{
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
	}
}
