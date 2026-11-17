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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
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
	tfClient       terraform.Client
	queueClient    *queueclient.Client
	featuresClient *features.Features
}

var _ apiPkg.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// show all components for an org
	apiAccess := require.Routes(api, permissions.APICtl)
	apiAccess.GET("/v1/components", require.Collection(permissions.ComponentInspect, s.componentResources), s.GetOrgComponents)
	apiAccess.GET("/v1/component-builds", require.Collection(permissions.ComponentBuildInspect, s.buildResources), s.ListOrgComponentBuilds)

	// components belong to an app
	apps := api.Group("/v1/apps/:app_id", require.AppInOrg(s.db))
	{
		components := require.Routes(apps.Group("/components"), permissions.APICtl)
		{
			components.GET("", require.Collection(permissions.ComponentInspect, s.componentResources), s.GetAppComponents)
			components.GET("/label-keys", require.Collection(permissions.ComponentInspect, s.componentResources), s.GetComponentLabelKeys)
			components.POST("", require.Point(permissions.ComponentCreate, s.componentAppTarget), s.CreateComponent)
			components.POST("/build-all", require.All(s.buildAllRequirements), s.BuildAllComponents)
		}

		// single component routes
		component := require.Routes(apps.Group("/component"), permissions.APICtl)
		{
			component.GET("/:component_name_or_id", require.Point(permissions.ComponentInspect, s.namedComponentTarget), s.GetAppComponent)
		}

		// component-specific routes
		comp := apps.Group("/components/:component_id")
		compAccess := require.Routes(comp, permissions.APICtl)
		{
			compAccess.PATCH("", require.Point(permissions.ComponentUpdate, s.componentTarget), s.UpdateAppComponent)
			compAccess.DELETE("", require.Point(permissions.ComponentDelete, s.componentTarget), s.DeleteAppComponent)
			compAccess.POST("/labels", require.Point(permissions.ComponentUpdate, s.componentTarget), s.AddAppComponentLabels)
			compAccess.DELETE("/labels", require.Point(permissions.ComponentUpdate, s.componentTarget), s.RemoveAppComponentLabels)
			// dependencies
			compAccess.GET("/dependencies", require.Point(permissions.AppConfigRead, s.componentGraphTarget), s.GetAppComponentDependencies)
			compAccess.GET("/dependents", require.Point(permissions.AppConfigRead, s.componentGraphTarget), s.GetAppComponentDependents)

			// builds
			builds := require.Routes(comp.Group("/builds"), permissions.APICtl)
			{
				builds.POST("", require.Point(permissions.ComponentBuildCreate, s.namedComponentTarget), s.CreateAppComponentBuild)
				builds.GET("/latest", require.Collection(permissions.ComponentBuildInspect, s.buildResources), s.GetAppComponentLatestBuild)
				builds.GET("/:build_id", require.Point(permissions.ComponentBuildInspect, s.buildTarget), s.GetAppComponentBuild)
				builds.POST("/:build_id/cancel", require.Point(permissions.ComponentBuildCancel, s.buildTarget), s.CancelAppComponentBuild)
				builds.GET("", require.Collection(permissions.ComponentBuildInspect, s.buildResources), s.GetAppComponentBuilds)
			}

			// component configurations
			configs := require.Routes(comp.Group("/configs"), permissions.APICtl)
			{
				configs.POST("/terraform-module", require.All(s.componentConfigRequirements), s.CreateAppTerraformModuleComponentConfig)
				configs.POST("/helm", require.All(s.componentConfigRequirements), s.CreateAppHelmComponentConfig)
				configs.POST("/external-image", require.All(s.componentConfigRequirements), s.CreateAppExternalImageComponentConfig)
				configs.POST("/job", require.All(s.componentConfigRequirements), s.CreateAppJobComponentConfig)
				configs.POST("/kubernetes-manifest", require.All(s.componentConfigRequirements), s.CreateAppKubernetesManifestComponentConfig)
				configs.POST("/pulumi", require.All(s.componentConfigRequirements), s.CreateAppPulumiComponentConfig)
				configs.GET("", require.Point(permissions.ComponentConfigRead, s.componentTarget), s.GetAppComponentConfigs)
				configs.GET("/:config_id", require.Point(permissions.ComponentConfigRead, s.componentTarget), s.GetAppComponentConfig)
				configs.GET("/latest", require.Point(permissions.ComponentConfigRead, s.componentTarget), s.GetAppComponentLatestConfig)
			}
		}
	}

	// deprecated routes
	deprecatedComponents := api.Group("/v1/components/:component_id", require.ComponentInOrg(s.db))
	// crud ops for components
	deprecatedComponentsAccess := require.Routes(deprecatedComponents, permissions.APICtl)
	{

		s.GET(deprecatedComponentsAccess.For(require.Point(permissions.ComponentInspect, s.namedComponentTarget)), "", s.GetComponent, apiPkg.APIContextTypePublic, true)
		s.PATCH(deprecatedComponentsAccess.For(require.Point(permissions.ComponentUpdate, s.componentTarget)), "", s.UpdateComponent, apiPkg.APIContextTypePublic, true)
		s.DELETE(deprecatedComponentsAccess.For(require.Point(permissions.ComponentDelete, s.componentTarget)), "", s.DeleteComponent, apiPkg.APIContextTypePublic, true)

		// dependencies
		s.GET(deprecatedComponentsAccess.For(require.Point(permissions.AppConfigRead, s.componentGraphTarget)), "/dependencies", s.GetComponentDependencies, apiPkg.APIContextTypePublic, true)
		s.GET(deprecatedComponentsAccess.For(require.Point(permissions.AppConfigRead, s.componentGraphTarget)), "/dependents", s.GetComponentDependents, apiPkg.APIContextTypePublic, true)

		// component configurations
		deprecatedConfigs := require.Routes(deprecatedComponents.Group("/configs"), permissions.APICtl)
		{
			s.POST(deprecatedConfigs.For(require.All(s.componentConfigRequirements)), "/terraform-module", s.CreateTerraformModuleComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs.For(require.All(s.componentConfigRequirements)), "/helm", s.CreateHelmComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs.For(require.All(s.componentConfigRequirements)), "/external-image", s.CreateExternalImageComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs.For(require.All(s.componentConfigRequirements)), "/job", s.CreateJobComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs.For(require.All(s.componentConfigRequirements)), "/kubernetes-manifest", s.CreateKubernetesManifestComponentConfig, apiPkg.APIContextTypePublic, true)
			s.POST(deprecatedConfigs.For(require.All(s.componentConfigRequirements)), "/pulumi", s.CreatePulumiComponentConfig, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedConfigs.For(require.Point(permissions.ComponentConfigRead, s.componentTarget)), "", s.GetComponentConfigs, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedConfigs.For(require.Point(permissions.ComponentConfigRead, s.componentTarget)), "/:config_id", s.GetComponentConfig, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedConfigs.For(require.Point(permissions.ComponentConfigRead, s.componentTarget)), "/latest", s.GetComponentLatestConfig, apiPkg.APIContextTypePublic, true)
		}

		// builds
		deprecatedBuilds := require.Routes(deprecatedComponents.Group("/builds"), permissions.APICtl)
		{
			s.POST(deprecatedBuilds.For(require.Point(permissions.ComponentBuildCreate, s.namedComponentTarget)), "", s.CreateComponentBuild, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedBuilds.For(require.Collection(permissions.ComponentBuildInspect, s.buildResources)), "/latest", s.GetComponentLatestBuild, apiPkg.APIContextTypePublic, true)
			s.GET(deprecatedBuilds.For(require.Point(permissions.ComponentBuildInspect, s.buildTarget)), "/:build_id", s.GetComponentBuild, apiPkg.APIContextTypePublic, true)
		}
	}

	// other deprecated build routes
	s.GET(apiAccess.For(require.Collection(permissions.ComponentBuildInspect, s.buildResources)), "/v1/builds", s.GetComponentBuilds, apiPkg.APIContextTypePublic, true)
	s.GET(apiAccess.For(require.Point(permissions.ComponentBuildInspect, s.buildTarget)), "/v1/components/builds/:build_id", s.GetBuild, apiPkg.APIContextTypePublic, true)

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
		tfClient:       params.TfClient,
		queueClient:    params.QueueClient,
		featuresClient: params.FeaturesClient,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
