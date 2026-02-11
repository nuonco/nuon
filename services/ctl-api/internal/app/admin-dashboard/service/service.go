package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In
	V           *validator.Validate
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	MW          metrics.Writer
	L           *zap.Logger
	AppsHelpers *appshelpers.Helpers
}

type Service struct {
	v           *validator.Validate
	l           *zap.Logger
	db          *gorm.DB
	mw          metrics.Writer
	cfg         *internal.Config
	appsHelpers *appshelpers.Helpers
}

type service = Service

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
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
	// Serve static assets
	api.Static("/assets", "./internal/app/admin-dashboard/assets")

	// Register routes - templ components will be rendered directly in handlers
	api.GET("/", s.Index)
	api.GET("/orgs", s.Orgs)
	api.GET("/orgs/table", s.OrgsTable)
	api.GET("/orgs/:id", s.OrgDetail)
	api.GET("/orgs/:id/status", s.OrgStatus)
	api.POST("/orgs/:id/tags", s.UpdateOrgTags)
	api.POST("/orgs/:id/tags/remove/:tag", s.RemoveSingleTag)
	api.GET("/orgs/:id/installs/table", s.InstallsTable)

	// Accounts routes
	api.GET("/accounts", s.Accounts)
	api.GET("/accounts/table", s.AccountsTable)
	api.GET("/accounts/:id", s.AccountDetail)
	api.GET("/accounts/:id/installs/table", s.AccountInstallsTable)
	api.GET("/accounts/:id/audit-logs/table", s.AccountAuditLogsTable)

	// Global installs routes
	api.GET("/installs", s.Installs)
	api.GET("/installs/table", s.InstallsTableGlobal)

	// Install detail routes
	api.GET("/installs/:id", s.InstallDetail)
	api.GET("/installs/:id/status/runner", s.InstallRunnerStatus)
	api.GET("/installs/:id/status/sandbox", s.InstallSandboxStatus)
	api.GET("/installs/:id/status/component", s.InstallComponentStatus)
	api.GET("/installs/:id/status/drift", s.InstallDriftStatus)

	s.l.Info("admin-dashboard routes registered")
	return nil
}

func New(params Params) (*service, error) {
	s := &service{
		cfg:         params.Cfg,
		l:           params.L,
		v:           params.V,
		db:          params.DB,
		mw:          params.MW,
		appsHelpers: params.AppsHelpers,
	}

	s.l.Info("admin-dashboard service initialized")
	return s, nil
}
