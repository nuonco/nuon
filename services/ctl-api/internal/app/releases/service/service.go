package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	componenthelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type service struct {
	v           *validator.Validate
	l           *zap.Logger
	db          *gorm.DB
	mw          metrics.Writer
	cfg         *internal.Config
	compHelpers *componenthelpers.Helpers
	evClient    eventloop.Client
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
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

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB,
	mw metrics.Writer,
	l *zap.Logger,
	compHelpers *componenthelpers.Helpers,
	evClient eventloop.Client,
) *service {
	return &service{
		cfg:         cfg,
		l:           l,
		v:           v,
		db:          db,
		mw:          mw,
		compHelpers: compHelpers,
		evClient:    evClient,
	}
}
