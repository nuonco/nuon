package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	V           *validator.Validate
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	MW          metrics.Writer
	L           *zap.Logger
	Helpers     *helpers.Helpers
	VcsHelpers  *vcshelpers.Helpers
	AppsHelpers *appshelpers.Helpers
	EvClient    eventloop.Client
}

type service struct {
	v           *validator.Validate
	l           *zap.Logger
	db          *gorm.DB
	mw          metrics.Writer
	cfg         *internal.Config
	helpers     *helpers.Helpers
	vcsHelpers  *vcshelpers.Helpers
	appsHelpers *appshelpers.Helpers
	evClient    eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// show all components for an org
	api.GET("/v1/components", s.GetOrgComponents)

	// components belong to an app
	api.GET("/v1/apps/:app_id/components", s.GetAppComponents)
	api.POST("/v1/apps/:app_id/components", s.CreateComponent)
	api.POST("/v1/apps/:app_id/components/build-all", s.BuildAllComponents)
	api.GET("/v1/apps/:app_id/component/:component_name_or_id", s.GetAppComponent)
	api.PATCH("/v1/apps/:app_id/components/:component_id", s.UpdateAppComponent)
	api.DELETE("/v1/apps/:app_id/components/:component_id", s.DeleteAppComponent)

	// get a component's dependencies
	api.GET("/v1/apps/:app_id/components/:component_id/dependencies", s.GetAppComponentDependencies)
	api.GET("/v1/apps/:app_id/components/:component_id/dependents", s.GetAppComponentDependents)

	// builds
	api.POST("/v1/apps/:app_id/components/:component_id/builds", s.CreateAppComponentBuild)
	api.GET("/v1/apps/:app_id/components/:component_id/builds/latest", s.GetAppComponentLatestBuild)
	api.GET("/v1/apps/:app_id/components/:component_id/builds/:build_id", s.GetAppComponentBuild)
	api.GET("/v1/apps/:app_id/components/:component_id/builds", s.GetAppComponentBuilds)

	// component configurations
	api.POST("/v1/apps/:app_id/components/:component_id/configs/terraform-module", s.CreateAppTerraformModuleComponentConfig)
	api.POST("/v1/apps/:app_id/components/:component_id/configs/helm", s.CreateAppHelmComponentConfig)
	api.POST("/v1/apps/:app_id/components/:component_id/configs/docker-build", s.CreateAppDockerBuildComponentConfig)
	api.POST("/v1/apps/:app_id/components/:component_id/configs/external-image", s.CreateAppExternalImageComponentConfig)
	api.POST("/v1/apps/:app_id/components/:component_id/configs/job", s.CreateAppJobComponentConfig)
	api.POST("/v1/apps/:app_id/components/:component_id/configs/kubernetes-manifest", s.CreateAppKubernetesManifestComponentConfig)

	api.GET("/v1/apps/:app_id/components/:component_id/configs", s.GetAppComponentConfigs)
	api.GET("/v1/apps/:app_id/components/:component_id/configs/:config_id", s.GetAppComponentConfig)
	api.GET("/v1/apps/:app_id/components/:component_id/configs/latest", s.GetAppComponentLatestConfig)

	// deprecated routes

	// crud ops for components
	api.GET("/v1/components/:component_id", s.GetComponent)       // Deprecated
	api.PATCH("/v1/components/:component_id", s.UpdateComponent)  // Deprecated
	api.DELETE("/v1/components/:component_id", s.DeleteComponent) // Deprecated

	// get a component's dependencies
	api.GET("/v1/components/:component_id/dependencies", s.GetComponentDependencies) // Deprecated
	api.GET("/v1/components/:component_id/dependents", s.GetComponentDependents)     // Deprecated

	// create component configurations
	api.POST("/v1/components/:component_id/configs/terraform-module", s.CreateTerraformModuleComponentConfig)       // Deprecated
	api.POST("/v1/components/:component_id/configs/helm", s.CreateHelmComponentConfig)                              // Deprecated
	api.POST("/v1/components/:component_id/configs/docker-build", s.CreateDockerBuildComponentConfig)               // Deprecated
	api.POST("/v1/components/:component_id/configs/external-image", s.CreateExternalImageComponentConfig)           // Deprecated
	api.POST("/v1/components/:component_id/configs/job", s.CreateJobComponentConfig)                                // Deprecated
	api.POST("/v1/components/:component_id/configs/kubernetes-manifest", s.CreateKubernetesManifestComponentConfig) // Deprecated
	api.GET("/v1/components/:component_id/configs", s.GetComponentConfigs)                                          // Deprecated
	api.GET("/v1/components/:component_id/configs/:config_id", s.GetComponentConfig)                                // Deprecated
	api.GET("/v1/components/:component_id/configs/latest", s.GetComponentLatestConfig)

	// builds are immutable
	api.POST("/v1/components/:component_id/builds", s.CreateComponentBuild)          // Deprecated
	api.GET("/v1/components/:component_id/builds/latest", s.GetComponentLatestBuild) // Deprecated
	api.GET("/v1/components/:component_id/builds/:build_id", s.GetComponentBuild)    // Deprecated

	api.GET("/v1/builds", s.GetComponentBuilds)            // deprecated
	api.GET("/v1/components/builds/:build_id", s.GetBuild) // deprecated

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/components", s.GetAllComponents)
	api.POST("/v1/components/:component_id/admin-restart", s.RestartComponent)
	api.POST("/v1/components/:component_id/admin-delete", s.AdminDeleteComponent)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		cfg:         params.Cfg,
		l:           params.L,
		v:           params.V,
		db:          params.DB,
		mw:          params.MW,
		helpers:     params.Helpers,
		vcsHelpers:  params.VcsHelpers,
		appsHelpers: params.AppsHelpers,
		evClient:    params.EvClient,
	}
}
