package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	installhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
)

type Params struct {
	fx.In

	V              *validator.Validate
	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	MW             metrics.Writer
	InstallHelpers *installhelpers.Helpers
	AuthzClient    *authz.Client
}

type service struct {
	v              *validator.Validate
	db             *gorm.DB
	installHelpers *installhelpers.Helpers
	authzClient    *authz.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
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

	api.POST("/v1/installers/hosted-installer-service-account", s.CreateHostedInstallerServiceAccount)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		v:              params.V,
		db:             params.DB,
		installHelpers: params.InstallHelpers,
		authzClient:    params.AuthzClient,
	}
}
