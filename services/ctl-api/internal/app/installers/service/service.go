package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	installhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	installhooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	"gorm.io/gorm"
)

type service struct {
	v              *validator.Validate
	db             *gorm.DB
	installHelpers *installhelpers.Helpers
	installHooks   *installhooks.Hooks
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// public installer endpoints
	api.GET("/v1/installer/:installer_slug/render", s.RenderAppInstaller)
	api.POST("/v1/installer/:installer_slug/installs", s.CreateInstallerInstall)
	api.GET("/v1/installer/:installer_slug/install/:install_id", s.GetInstallerInstall)

	// installers
	api.POST("/v1/installers", s.CreateAppInstaller)
	api.PATCH("/v1/installers/:installer_id", s.UpdateAppInstaller)
	api.DELETE("/v1/installers/:installer_id", s.DeleteAppInstaller)
	api.GET("/v1/installers/:installer_id", s.GetAppInstaller)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/installers", s.GetAllAppInstallers)

	return nil
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB, mw metrics.Writer,
	installHooks *installhooks.Hooks,
	installHelpers *installhelpers.Helpers,
) *service {
	return &service{
		v:              v,
		db:             db,
		installHelpers: installHelpers,
		installHooks:   installHooks,
	}
}
