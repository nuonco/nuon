package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	installhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
)

type service struct {
	v              *validator.Validate
	db             *gorm.DB
	installHelpers *installhelpers.Helpers
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// installers
	api.GET("/v1/installers", s.GetInstallers)
	api.POST("/v1/installers", s.CreateInstaller)

	api.PATCH("/v1/installers/:installer_id", s.UpdateInstaller)
	api.DELETE("/v1/installers/:installer_id", s.DeleteInstaller)
	api.GET("/v1/installers/:installer_id", s.GetInstaller)

	api.GET("/v1/installers/:installer_id/render", s.RenderAppInstaller)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/installers", s.GetAllInstallers)

	return nil
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB, mw metrics.Writer,
	installHelpers *installhelpers.Helpers,
) *service {
	return &service{
		v:              v,
		db:             db,
		installHelpers: installHelpers,
	}
}
