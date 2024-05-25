package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/multi"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/terraformcloud"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/hooks"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v           *validator.Validate
	l           *zap.Logger
	db          *gorm.DB
	mw          metrics.Writer
	cfg         *internal.Config
	hooks       *hooks.Hooks
	orgsOutputs *terraformcloud.OrgsOutputs
	wpClient    multi.Client
	helpers     *helpers.Helpers
	vcsHelpers  *vcshelpers.Helpers
	appsHelpers *appshelpers.Helpers
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// show all components for an org
	api.GET("/v1/components", s.GetOrgComponents)

	// components belong to an app
	api.GET("/v1/apps/:app_id/components", s.GetAppComponents)
	api.POST("/v1/apps/:app_id/components", s.CreateComponent)

	// crud ops for components
	api.GET("/v1/components/:component_id", s.GetComponent)
	api.PATCH("/v1/components/:component_id", s.UpdateComponent)
	api.DELETE("/v1/components/:component_id", s.DeleteComponent)

	// create component configurations
	api.POST("/v1/components/:component_id/configs/terraform-module", s.CreateTerraformModuleComponentConfig)
	api.POST("/v1/components/:component_id/configs/helm", s.CreateHelmComponentConfig)
	api.POST("/v1/components/:component_id/configs/docker-build", s.CreateDockerBuildComponentConfig)
	api.POST("/v1/components/:component_id/configs/external-image", s.CreateExternalImageComponentConfig)
	api.POST("/v1/components/:component_id/configs/job", s.CreateJobComponentConfig)
	api.GET("/v1/components/:component_id/configs", s.GetComponentConfigs)
	api.GET("/v1/components/:component_id/configs/latest", s.GetComponentLatestConfig)

	// builds are immutable
	api.POST("/v1/components/:component_id/builds", s.CreateComponentBuild)
	api.GET("/v1/components/:component_id/builds/latest", s.GetComponentLatestBuild)
	api.GET("/v1/components/:component_id/builds/:build_id", s.GetComponentBuild)
	api.GET("/v1/components/:component_id/builds/:build_id/logs", s.GetComponentBuildLogs)
	api.GET("/v1/components/:component_id/builds/:build_id/plan", s.GetComponentBuildPlan)

	api.GET("/v1/builds", s.GetComponentBuilds)
	api.GET("/v1/components/builds/:build_id", s.GetBuild)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/components", s.GetAllComponents)
	api.POST("/v1/components/:component_id/admin-restart", s.RestartComponent)
	api.POST("/v1/components/:component_id/admin-delete", s.AdminDeleteComponent)
	return nil
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB,
	mw metrics.Writer,
	l *zap.Logger,
	hooks *hooks.Hooks,
	orgsOutputs *terraformcloud.OrgsOutputs,
	wpClient multi.Client,
	helpers *helpers.Helpers,
	vcsHelpers *vcshelpers.Helpers,
	appsHelpers *appshelpers.Helpers,
) *service {
	return &service{
		cfg:         cfg,
		l:           l,
		v:           v,
		db:          db,
		mw:          mw,
		hooks:       hooks,
		orgsOutputs: orgsOutputs,
		wpClient:    wpClient,
		helpers:     helpers,
		vcsHelpers:  vcsHelpers,
		appsHelpers: appsHelpers,
	}
}
