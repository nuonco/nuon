package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	installhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
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
	AcctClient     *account.Client
}

type service struct {
	v              *validator.Validate
	db             *gorm.DB
	installHelpers *installhelpers.Helpers
	authzClient    *authz.Client
	acctClient     *account.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// installers
	installers := api.Group("/v1/installers")
	{
		installers.GET("", s.GetInstallers)
		installers.POST("", s.CreateInstaller)

		installer := installers.Group("/:installer_id")
		{
			installer.PATCH("", s.UpdateInstaller)
			installer.DELETE("", s.DeleteInstaller)
			installer.GET("", s.GetInstaller)
			installer.GET("/render", s.RenderAppInstaller)
		}
	}

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
		acctClient:     params.AcctClient,
	}
}
