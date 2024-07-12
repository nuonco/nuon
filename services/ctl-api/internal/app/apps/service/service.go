package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type service struct {
	v          *validator.Validate
	db         *gorm.DB
	mw         metrics.Writer
	cfg        *internal.Config
	vcsHelpers *vcshelpers.Helpers
	helpers    *appshelpers.Helpers
	evClient   eventloop.Client
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// manage apps
	api.POST("/v1/apps", s.CreateApp)
	api.GET("/v1/apps", s.GetApps)
	api.PATCH("/v1/apps/:app_id", s.UpdateApp)
	api.GET("/v1/apps/:app_id", s.GetApp)
	api.DELETE("/v1/apps/:app_id", s.DeleteApp)

	// app configs
	api.GET("/v1/apps/:app_id/template-config", s.GetAppConfigTemplate)
	api.POST("/v1/apps/:app_id/config", s.CreateAppConfig)
	api.GET("/v1/apps/:app_id/configs", s.GetAppConfigs)
	api.GET("/v1/apps/:app_id/config/:app_config_id", s.GetAppConfig)
	api.GET("/v1/apps/:app_id/latest-config", s.GetAppLatestConfig)
	api.POST("/v1/apps/:app_id/config/:app_config_id/set-status", s.SetAppConfigStatus)

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

	// app secrets management
	api.POST("/v1/apps/:app_id/secret", s.CreateAppSecret)
	api.GET("/v1/apps/:app_id/secrets", s.GetAppSecrets)
	api.DELETE("/v1/apps/:app_id/secret/:secret_id", s.DeleteAppSecret)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/apps", s.GetAllApps)
	api.POST("/v1/apps/:app_id/admin-reprovision", s.AdminReprovisionApp)
	api.POST("/v1/apps/:app_id/admin-restart", s.RestartApp)
	api.POST("/v1/apps/:app_id/admin-delete", s.AdminDeleteApp)
	api.POST("/v1/apps/:app_id/admin-aws-delegation", s.AdminAddAWSDelegationConfig)

	return nil
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB, mw metrics.Writer,
	ghClient *github.Client,
	vcsHelpers *vcshelpers.Helpers,
	helpers *appshelpers.Helpers,
	evClient eventloop.Client,
) *service {
	return &service{
		cfg:        cfg,
		v:          v,
		db:         db,
		mw:         mw,
		vcsHelpers: vcsHelpers,
		helpers:    helpers,
		evClient:   evClient,
	}
}
