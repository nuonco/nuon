package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	componenthelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	V           *validator.Validate
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	MW          metrics.Writer
	L           *zap.Logger
	CompHelpers *componenthelpers.Helpers
	EvClient    eventloop.Client
}

type service struct {
	v           *validator.Validate
	l           *zap.Logger
	db          *gorm.DB
	mw          metrics.Writer
	cfg         *internal.Config
	compHelpers *componenthelpers.Helpers
	evClient    eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	api.POST("/v1/components/:component_id/releases", s.CreateComponentRelease)
	api.GET("/v1/components/:component_id/releases", s.GetComponentReleases)
	api.GET("/v1/apps/:app_id/releases", s.GetAppReleases)

	api.GET("/v1/releases/:release_id", s.GetRelease)
	api.GET("/v1/releases/:release_id/steps", s.GetReleaseSteps)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/releases", s.GetAllReleases)
	api.POST("/v1/releases/:release_id/admin-restart", s.RestartRelease)
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		cfg:         params.Cfg,
		l:           params.L,
		v:           params.V,
		db:          params.DB,
		mw:          params.MW,
		compHelpers: params.CompHelpers,
		evClient:    params.EvClient,
	}
}
