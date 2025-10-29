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
	apps := api.Group("/v1/apps/:app_id")
	{
		components := apps.Group("/components")
		{
			components.GET("", s.GetAppComponents)
			components.POST("", s.CreateComponent)
			components.POST("/build-all", s.BuildAllComponents)
		}

		// single component routes
		component := apps.Group("/component")
		{
			component.GET("/:component_name_or_id", s.GetAppComponent)
		}

		// component-specific routes
		comp := apps.Group("/components/:component_id")
		{
			comp.PATCH("", s.UpdateAppComponent)
			comp.DELETE("", s.DeleteAppComponent)

			// dependencies
			comp.GET("/dependencies", s.GetAppComponentDependencies)
			comp.GET("/dependents", s.GetAppComponentDependents)

			// builds
			builds := comp.Group("/builds")
			{
				builds.POST("", s.CreateAppComponentBuild)
				builds.GET("/latest", s.GetAppComponentLatestBuild)
				builds.GET("/:build_id", s.GetAppComponentBuild)
				builds.GET("", s.GetAppComponentBuilds)
			}

			// component configurations
			configs := comp.Group("/configs")
			{
				configs.POST("/terraform-module", s.CreateAppTerraformModuleComponentConfig)
				configs.POST("/helm", s.CreateAppHelmComponentConfig)
				configs.POST("/docker-build", s.CreateAppDockerBuildComponentConfig)
				configs.POST("/external-image", s.CreateAppExternalImageComponentConfig)
				configs.POST("/job", s.CreateAppJobComponentConfig)
				configs.POST("/kubernetes-manifest", s.CreateAppKubernetesManifestComponentConfig)
				configs.GET("", s.GetAppComponentConfigs)
				configs.GET("/:config_id", s.GetAppComponentConfig)
				configs.GET("/latest", s.GetAppComponentLatestConfig)
			}
		}
	}

	// deprecated routes
	deprecatedComponents := api.Group("/v1/components/:component_id")
	{
		// crud ops for components
		deprecatedComponents.GET("", s.GetComponent)       // Deprecated
		deprecatedComponents.PATCH("", s.UpdateComponent)  // Deprecated
		deprecatedComponents.DELETE("", s.DeleteComponent) // Deprecated

		// dependencies
		deprecatedComponents.GET("/dependencies", s.GetComponentDependencies) // Deprecated
		deprecatedComponents.GET("/dependents", s.GetComponentDependents)     // Deprecated

		// component configurations
		deprecatedConfigs := deprecatedComponents.Group("/configs")
		{
			deprecatedConfigs.POST("/terraform-module", s.CreateTerraformModuleComponentConfig)       // Deprecated
			deprecatedConfigs.POST("/helm", s.CreateHelmComponentConfig)                              // Deprecated
			deprecatedConfigs.POST("/docker-build", s.CreateDockerBuildComponentConfig)               // Deprecated
			deprecatedConfigs.POST("/external-image", s.CreateExternalImageComponentConfig)           // Deprecated
			deprecatedConfigs.POST("/job", s.CreateJobComponentConfig)                                // Deprecated
			deprecatedConfigs.POST("/kubernetes-manifest", s.CreateKubernetesManifestComponentConfig) // Deprecated
			deprecatedConfigs.GET("", s.GetComponentConfigs)                                          // Deprecated
			deprecatedConfigs.GET("/:config_id", s.GetComponentConfig)                                // Deprecated
			deprecatedConfigs.GET("/latest", s.GetComponentLatestConfig)                              // Deprecated
		}

		// builds
		deprecatedBuilds := deprecatedComponents.Group("/builds")
		{
			deprecatedBuilds.POST("", s.CreateComponentBuild)          // Deprecated
			deprecatedBuilds.GET("/latest", s.GetComponentLatestBuild) // Deprecated
			deprecatedBuilds.GET("/:build_id", s.GetComponentBuild)    // Deprecated
		}
	}

	// other deprecated build routes
	api.GET("/v1/builds", s.GetComponentBuilds)            // deprecated
	api.GET("/v1/components/builds/:build_id", s.GetBuild) // deprecated

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	components := api.Group("/v1/components")
	{
		components.GET("", s.GetAllComponents)

		// component admin routes
		component := components.Group("/:component_id")
		{
			component.POST("/admin-restart", s.RestartComponent)
			component.POST("/admin-delete", s.AdminDeleteComponent)
		}
	}

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
