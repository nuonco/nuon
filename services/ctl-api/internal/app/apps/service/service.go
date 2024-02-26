package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/hooks"
	componenthooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/hooks"
	installhooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"gorm.io/gorm"
)

type service struct {
	v              *validator.Validate
	db             *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	hooks          *hooks.Hooks
	installHooks   *installhooks.Hooks
	componentHooks *componenthooks.Hooks
	vcsHelpers     *vcshelpers.Helpers
	helpers        *appshelpers.Helpers
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// manage apps
	api.POST("/v1/apps", s.CreateApp)
	api.GET("/v1/apps", s.GetApps)
	api.PATCH("/v1/apps/:app_id", s.UpdateApp)
	api.GET("/v1/apps/:app_id", s.GetApp)
	api.DELETE("/v1/apps/:app_id", s.DeleteApp)

	// app configs
	api.POST("/v1/apps/:app_id/config", s.CreateAppConfig)
	api.GET("/v1/apps/:app_id/latest-config", s.GetAppLatestConfig)
	api.GET("/v1/apps/:app_id/configs", s.GetAppConfigs)

	// app sandbox management
	api.POST("/v1/apps/:app_id/sandbox-config", s.CreateAppSandboxConfig)
	api.GET("/v1/apps/:app_id/sandbox-latest-config", s.GetAppSandboxLatestConfig)
	api.GET("/v1/apps/:app_id/sandbox-configs", s.GetAppSandboxConfigs)

	// app runner management
	api.POST("/v1/apps/:app_id/runner-config", s.CreateAppRunnerConfig)
	api.GET("/v1/apps/:app_id/runner-latest-config", s.GetAppRunnerLatestConfig)
	api.GET("/v1/apps/:app_id/runner-configs", s.GetAppRunnerConfigs)

	// app input management
	api.POST("/v1/apps/:app_id/input-config", s.CreateAppInputsConfig)
	api.GET("/v1/apps/:app_id/input-latest-config", s.GetAppInputLatestConfig)
	api.GET("/v1/apps/:app_id/input-configs", s.GetAppInputConfigs)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/apps", s.GetAllApps)
	api.POST("/v1/apps/:app_id/admin-reprovision", s.AdminReprovisionApp)
	api.POST("/v1/apps/:app_id/admin-restart", s.RestartApp)
	api.POST("/v1/apps/:app_id/admin-delete", s.AdminDeleteApp)

	return nil
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB, mw metrics.Writer,
	hooks *hooks.Hooks,
	installHooks *installhooks.Hooks,
	componentHooks *componenthooks.Hooks,
	ghClient *github.Client,
	vcsHelpers *vcshelpers.Helpers,
	helpers *appshelpers.Helpers,
) *service {
	return &service{
		cfg:            cfg,
		v:              v,
		db:             db,
		mw:             mw,
		hooks:          hooks,
		installHooks:   installHooks,
		componentHooks: componentHooks,
		vcsHelpers:     vcsHelpers,
		helpers:        helpers,
	}
}
