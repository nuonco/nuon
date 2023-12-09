package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/hooks"
	componenthooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/hooks"
	installhooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	hooks          *hooks.Hooks
	installHooks   *installhooks.Hooks
	componentHooks *componenthooks.Hooks
	ghClient       *github.Client
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// public installer endpoints
	api.GET("/v1/installer/:installer_slug/render", s.RenderAppInstaller)
	api.POST("/v1/installer/:installer_slug/installs", s.CreateInstallerInstall)
	api.GET("/v1/installer/:installer_slug/install/:install_id", s.GetInstallerInstall)

	// manage apps
	api.POST("/v1/apps", s.CreateApp)
	api.GET("/v1/apps", s.GetApps)
	api.PATCH("/v1/apps/:app_id", s.UpdateApp)
	api.GET("/v1/apps/:app_id", s.GetApp)
	api.DELETE("/v1/apps/:app_id", s.DeleteApp)

	// app sandbox management
	api.POST("/v1/apps/:app_id/sandbox-config", s.CreateAppSandboxConfig)
	api.GET("/v1/apps/:app_id/sandbox-latest-config", s.GetAppSandboxLatestConfig)
	api.GET("/v1/apps/:app_id/sandbox-configs", s.GetAppSandboxConfigs)

	// app input management
	api.POST("/v1/apps/:app_id/input-config", s.CreateAppInputsConfig)
	api.GET("/v1/apps/:app_id/input-latest-config", s.GetAppInputLatestConfig)
	api.GET("/v1/apps/:app_id/input-configs", s.GetAppInputConfigs)

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
	api.POST("/v1/apps/:app_id/admin-delete", s.AdminDeleteApp)

	api.GET("/v1/installers", s.GetAllAppInstallers)

	// temporary migrations

	// ensure all apps have an app sandbox config
	api.POST("/v1/apps/admin-add-app-sandbox-configs", s.AdminAddAppSandboxConfigs)

	// ensure all apps have an app input config
	api.POST("/v1/apps/admin-add-app-inputs", s.AdminAddAppInputsConfigs)

	// ensure all app sandbox configs have an app-id field
	api.POST("/v1/apps/admin-update-app-sandbox-configs", s.AdminUpdateAppSandboxConfigs)

	return nil
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB, mw metrics.Writer,
	l *zap.Logger,
	hooks *hooks.Hooks,
	installHooks *installhooks.Hooks,
	componentHooks *componenthooks.Hooks,
	ghClient *github.Client,
) *service {
	return &service{
		cfg:            cfg,
		l:              l,
		v:              v,
		db:             db,
		mw:             mw,
		hooks:          hooks,
		installHooks:   installHooks,
		componentHooks: componentHooks,
		ghClient:       ghClient,
	}
}
