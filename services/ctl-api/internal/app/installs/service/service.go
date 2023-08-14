package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/hooks"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v     *validator.Validate
	l     *zap.Logger
	db    *gorm.DB
	mw    metrics.Writer
	cfg   *internal.Config
	hooks *hooks.Hooks
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/v1/apps/:app_id/installs", s.GetAppInstalls)
	api.POST("/v1/apps/:app_id/installs", s.CreateInstall)

	api.GET("/v1/installs", s.GetOrgInstalls)

	api.GET("/v1/installs/:app_id/:install_id", s.GetInstall)
	api.PATCH("/v1/installs/:app_id/:install_id", s.UpdateInstall)
	api.DELETE("/v1/installs/:app_id/:install_id", s.DeleteInstall)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/installs", s.GetAllInstalls)
	return nil
}

func New(v *validator.Validate, cfg *internal.Config, db *gorm.DB, mw metrics.Writer, l *zap.Logger, hooks *hooks.Hooks) *service {
	return &service{
		cfg:   cfg,
		l:     l,
		v:     v,
		db:    db,
		mw:    mw,
		hooks: hooks,
	}
}
