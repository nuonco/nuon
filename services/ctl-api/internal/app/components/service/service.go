package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
)

type Params struct {
	fx.In

	V              *validator.Validate
	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	MW             metrics.Writer
	L              *zap.Logger
	Helpers        *helpers.Helpers
	VcsHelpers     *vcshelpers.Helpers
	AppsHelpers    *appshelpers.Helpers
	EvClient       eventloop.Client
	TfClient       terraform.Client
	QueueClient    *queueclient.Client
	FeaturesClient *features.Features
	EndpointAudit  *apiPkg.EndpointAudit
}

type service struct {
	apiPkg.RouteRegister
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	helpers        *helpers.Helpers
	vcsHelpers     *vcshelpers.Helpers
	appsHelpers    *appshelpers.Helpers
	evClient       eventloop.Client
	tfClient       terraform.Client
	queueClient    *queueclient.Client
	featuresClient *features.Features
}

var _ apiPkg.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// show all components for an org
	api.GET("/v1/components", s.GetOrgComponents)

	// components belong to an app
	apps := api.Group("/v1/apps/:app_id")
	{
		components := apps.Group("/components")
		{
			components.GET("", s.GetAppComponents)
			components.GET("/label-keys", s.GetComponentLabelKeys)
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
			comp.POST("/labels", s.AddAppComponentLabels)
			comp.DELETE("/labels", s.RemoveAppComponentLabels)

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
				configs.POST("/pulumi", s.CreateAppPulumiComponentConfig)
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
		s.GET(deprecatedComponents, "", s.GetComponent, apiPkg.APIContextTypePublic, true)
		s.PATCH(deprecatedComponents, "", s.UpdateComponent, apiPkg.APIContextTypePublic, true)
		s.DELETE(deprecatedComponents, "", s.DeleteComponent, apiPkg.APIContextTypePublic, true)

		// dependencies
		s.GET(deprecatedComponents, "/dependencies", s.GetComponentDependencies, apiPkg.APIContextTypePublic, true)
		s.GET(deprecatedComponents, "/dependents", s.GetComponentDependents, apiPkg.APIContextTypePublic, true)

		// component configurations
		deprecatedConfigs := deprecatedComponents.Group("/configs")
		{
			s.POST(deprecatedConfigs, "/terraform-module", s.CreateTerraformModuleComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs, "/helm", s.CreateHelmComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs, "/docker-build", s.CreateDockerBuildComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs, "/external-image", s.CreateExternalImageComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs, "/job", s.CreateJobComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs, "/kubernetes-manifest", s.CreateKubernetesManifestComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs, "/pulumi", s.CreatePulumiComponentConfig, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedConfigs, "", s.GetComponentConfigs, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedConfigs, "/:config_id", s.GetComponentConfig, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedConfigs, "/latest", s.GetComponentLatestConfig, apiPkg.APIContextTypePublic, true)
		}

		// builds
		deprecatedBuilds := deprecatedComponents.Group("/builds")
		{
			s.POST(deprecatedBuilds, "", s.CreateComponentBuild, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedBuilds, "/latest", s.GetComponentLatestBuild, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedBuilds, "/:build_id", s.GetComponentBuild, apiPkg.APIContextTypePublic, true)
		}
	}

	// other deprecated build routes
	s.GET(api, "/v1/builds", s.GetComponentBuilds, apiPkg.APIContextTypePublic, true)
	s.GET(api, "/v1/components/builds/:build_id", s.GetBuild, apiPkg.APIContextTypePublic, true)

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

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		RouteRegister: apiPkg.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		cfg:            params.Cfg,
		l:              params.L,
		v:              params.V,
		db:             params.DB,
		mw:             params.MW,
		helpers:        params.Helpers,
		vcsHelpers:     params.VcsHelpers,
		appsHelpers:    params.AppsHelpers,
		evClient:       params.EvClient,
		tfClient:       params.TfClient,
		queueClient:    params.QueueClient,
		featuresClient: params.FeaturesClient,
	}
}
