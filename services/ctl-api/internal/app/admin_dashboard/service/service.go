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
)

type Params struct {
	fx.In
	V   *validator.Validate
	Cfg *internal.Config
	DB  *gorm.DB `name:"psql"`
	MW  metrics.Writer
	L   *zap.Logger
}

type service struct {
	v   *validator.Validate
	l   *zap.Logger
	db  *gorm.DB
	mw  metrics.Writer
	cfg *internal.Config
}

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
	api.Static("/assets", "./internal/app/admin_dashboard/assets")

	// Register routes - templ components will be rendered directly in handlers
	api.GET("/", s.Index)
	api.GET("/livez", s.Livez)
	api.GET("/orgs", s.Orgs)
	api.GET("/orgs/table", s.OrgsTable)
	api.GET("/orgs/:id", s.OrgDetail)
	api.GET("/orgs/:id/status", s.OrgStatus)
	api.GET("/orgs/:id/installs/table", s.InstallsTable)

	s.l.Info("admin-dashboard routes registered")
	return nil
}

func New(params Params) (*service, error) {
	s := &service{
		cfg: params.Cfg,
		l:   params.L,
		v:   params.V,
		db:  params.DB,
		mw:  params.MW,
	}

	s.l.Info("admin-dashboard service initialized")
	return s, nil
}
