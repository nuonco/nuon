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
	// public installer endpoints
	api.GET("/v1/installer/:installer_slug/render", s.RenderAppInstaller)

	// manage apps
	api.POST("/v1/apps", s.CreateApp)
	api.GET("/v1/apps", s.GetApps)

	api.PATCH("/v1/apps/:app_id", s.UpdateApp)
	api.GET("/v1/apps/:app_id", s.GetApp)
	api.DELETE("/v1/apps/:app_id", s.DeleteApp)

	// sandbox release
	api.PUT("/v1/apps/:app_id/sandbox", s.UpdateAppSandbox)

	// installers
	api.POST("/v1/apps/:app_id/installer", s.CreateAppInstaller)
	api.PATCH("/v1/installers/:installer_id", s.UpdateAppInstaller)
	api.DELETE("/v1/installers/:installer_id", s.DeleteAppInstaller)
	api.GET("/v1/installers/:installer_id", s.GetAppInstaller)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/apps", s.GetAllApps)
	api.POST("/v1/apps/:app_id/admin-reprovision", s.AdminReprovisionApp)
	api.POST("/v1/apps/:app_id/admin-restart", s.RestartApp)
	api.POST("/v1/apps/:app_id/admin-update-sandbox", s.AdminUpdateSandbox)

	api.GET("/v1/installers", s.GetAllAppInstallers)
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
