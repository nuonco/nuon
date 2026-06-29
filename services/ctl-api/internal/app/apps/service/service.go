package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V               *validator.Validate
	DB              *gorm.DB `name:"psql"`
	MW              metrics.Writer
	L               *zap.Logger
	Cfg             *internal.Config
	VcsHelpers      *vcshelpers.Helpers
	Helpers         *appshelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	FeaturesClient  *features.Features
	QueueClient     *queueclient.Client
	EndpointAudit   *api.EndpointAudit
	TemporalClient  temporalclient.Client
}

type service struct {
	api.RouteRegister
	v               *validator.Validate
	db              *gorm.DB
	mw              metrics.Writer
	l               *zap.Logger
	cfg             *internal.Config
	vcsHelpers      *vcshelpers.Helpers
	helpers         *appshelpers.Helpers
	installsHelpers *installshelpers.Helpers
	accountsHelpers *accountshelpers.Helpers
	featuresClient  *features.Features
	temporalClient  temporalclient.Client
	queueClient     *queueclient.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	// manage apps
	apps := ge.Group("/v1/apps")
	{
		apps.POST("", s.CreateApp)
		apps.GET("", s.GetApps)
		apps.PATCH("/:app_id", s.UpdateApp)
		apps.GET("/:app_id", s.GetApp)
		apps.DELETE("/:app_id", s.DeleteApp)
	}

	// app-specific routes
	app := ge.Group("/v1/apps/:app_id")
	{
		// app configs
		app.GET("/template-config", s.GetAppConfigTemplate)
		appConfig := app.Group("/config") // deprecated singular route
		{
			s.POST(appConfig, "", s.CreateAppConfig, api.APIContextTypePublic, true)                                        // deprecated singular route
			s.GET(appConfig, "/:app_config_id", s.GetAppConfig, api.APIContextTypePublic, true)                             // deprecated singular route
			s.PATCH(appConfig, "/:app_config_id", s.UpdateAppConfig, api.APIContextTypePublic, true)                        // deprecated singular route
			s.POST(appConfig, "/:app_config_id/update-installs", s.UpdateAppConfigInstalls, api.APIContextTypePublic, true) // deprecated singular route
			s.GET(appConfig, "/:app_config_id/graph", s.GetAppConfigGraph, api.APIContextTypePublic, true)                  // deprecated singular route
		}

		appConfigs := app.Group("/configs")
		{
			appConfigs.GET("", s.GetAppConfigs)
			appConfigs.POST("", s.CreateAppConfigV2)
			appConfigs.PATCH("/:config_id", s.UpdateAppConfigV2)
			appConfigs.GET("/:config_id", s.GetAppConfigV2)
			appConfigs.POST("/:config_id/update-installs", s.UpdateAppConfigInstallsV2)
			appConfigs.GET("/:config_id/graph", s.GetAppConfigGraphV2)
			appConfigs.POST("/:config_id/build", s.BuildAppConfig)
			appConfigs.GET("/:config_id/diff", s.GetAppConfigDiff)
		}

		// app sandbox builds
		sandboxBuilds := app.Group("/sandbox/builds")
		{
			sandboxBuilds.GET("", s.GetAppSandboxBuilds)
			sandboxBuilds.POST("", s.CreateAppSandboxBuild)
			sandboxBuilds.GET("/:build_id", s.GetAppSandboxBuild)
		}

		// app sandbox management
		sandboxConfig := app.Group("/sandbox-config")
		{
			s.POST(sandboxConfig, "", s.CreateAppSandboxConfigV2, api.APIContextTypePublic, true) // deprecated singular route
		}

		sandboxConfigs := app.Group("/sandbox-configs")
		{
			sandboxConfigs.GET("", s.GetAppSandboxConfigs)
			sandboxConfigs.POST("", s.CreateAppSandboxConfigV2)
		}

		// app secrets configs management
		secretsConfigs := app.Group("/secrets-configs")
		{
			secretsConfigs.POST("", s.CreateAppSecretsConfig)
			secretsConfigs.GET("/:config_id", s.GetAppSecretsConfig)
		}

		// app stack configs
		stackConfigs := app.Group("/stack-configs")
		{
			stackConfigs.POST("", s.CreateAppStackConfig)
			stackConfigs.GET("/:config_id", s.GetAppSecretsConfig)
		}

		// app policies management
		policiesConfigs := app.Group("/policies-configs")
		{
			policiesConfigs.GET("", s.GetAppPoliciesConfigs)
			policiesConfigs.POST("", s.CreateAppPoliciesConfig)
			policiesConfigs.GET("/:config_id", s.GetAppPoliciesConfig)
		}

		// individual policy config lookup
		app.GET("/policy-config/:policy_config_id", s.GetAppPolicyConfig)

		// app break glass
		breakGlassConfigs := app.Group("/break-glass-configs")
		{
			breakGlassConfigs.POST("", s.CreateAppBreakGlasssConfig)
			breakGlassConfigs.GET("/:config_id", s.GetAppBreakGlassConfig)
		}

		// app permissions
		permissionsConfigs := app.Group("/permissions-configs")
		{
			permissionsConfigs.POST("", s.CreateAppPermissionsConfig)
			permissionsConfigs.GET("/:config_id", s.GetAppPermissionsConfig)
		}

		operationRoleConfigs := app.Group("/operation-role-configs")
		{
			operationRoleConfigs.POST("", s.CreateAppOperationRoleConfig)
			operationRoleConfigs.GET("/:operation_role_config_id", s.GetAppOperationRoleConfigs)
		}

		// app runner management
		runnerConfigs := app.Group("/runner-configs")
		{
			runnerConfigs.POST("", s.CreateAppRunnerConfig)
			runnerConfigs.GET("", s.GetAppRunnerConfigs)
		}

		// app input management
		app.POST("/input-config", s.CreateAppInputsConfig)
		inputConfigs := app.Group("/input-configs")
		{
			inputConfigs.GET("", s.GetAppInputConfigs)
			inputConfigs.GET("/:input_config_id", s.GetAppInputConfig)
		}

		// app secrets management
		s.POST(app, "/secret", s.CreateAppSecretV2, api.APIContextTypePublic, true) // deprecated singular route
		app.POST("/secrets", s.CreateAppSecretV2)
		secret := app.Group("/secret")
		{
			s.DELETE(secret, "/:secret_id", s.DeleteAppSecret, api.APIContextTypePublic, true) // deprecated singular route
		}

		secrets := app.Group("/secrets")
		{
			secrets.GET("", s.GetAppSecrets)
			secrets.DELETE("/:secret_id", s.DeleteAppSecretV2)

		}

		// app branches
		branches := app.Group("/branches")
		{
			branches.POST("", s.CreateAppBranch)
			branches.GET("", s.GetAppBranches)
			branches.GET("/:app_branch_id", s.GetAppBranch)
			branches.PATCH("/:app_branch_id", s.UpdateAppBranch)
			branches.DELETE("/:app_branch_id", s.DeleteAppBranch)
			branches.GET("/:app_branch_id/configs", s.GetAppBranchAppConfigs)
			branches.POST("/:app_branch_id/configs", s.CreateAppBranchConfig)
			branches.GET("/:app_branch_id/latest-config", s.GetAppBranchLatestConfig)
			branches.POST("/:app_branch_id/runs", s.TriggerAppBranchRun)
			branches.GET("/:app_branch_id/runs", s.GetAppBranchRuns)
			branches.GET("/:app_branch_id/runs/:run_id/builds", s.GetAppBranchRunBuilds)
			branches.GET("/:app_branch_id/runs/:run_id/install-groups", s.GetAppBranchRunInstallGroups)
			branches.GET("/:app_branch_id/runs/:run_id/install-group-runs", s.GetInstallGroupRuns)
			branches.GET("/:app_branch_id/runs/:run_id/install-group-runs/:install_group_run_id", s.GetInstallGroupRun)
		}

		// TODO deprecate - latest config routes
		app.GET("/latest-break-glass-config", s.GetLatestAppBreakGlassConfig)
		app.GET("/runner-latest-config", s.GetAppRunnerLatestConfig)
		app.GET("/input-latest-config", s.GetAppInputLatestConfig)
		app.GET("/latest-policies-config", s.GetLatestAppPoliciesConfig)
		app.GET("/latest-config", s.GetAppLatestConfig)
		app.GET("/sandbox-latest-config", s.GetAppSandboxLatestConfig)
		app.GET("/latest-secrets-config", s.GetLatestAppSecretsConfig)
		app.GET("/latest-permissions-config", s.GetLatestAppPermissionsConfig)
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// apps
	apps := api.Group("/v1/apps")
	{
		apps.GET("", s.GetAllApps)
		apps.GET("/details", s.AdminListAppsDetails)

		// app admin routes
		app := apps.Group("/:app_id")
		{
			app.POST("/admin-reprovision", s.AdminReprovisionApp)
			app.POST("/admin-restart", s.RestartApp)
			app.POST("/admin-config-graph", s.AdminConfigGraph)
		}
	}

	// app branches
	appBranches := api.Group("/v1/app-branches")
	{
		appBranches.POST("/:app_branch_id/admin-trigger-run", s.AdminTriggerAppBranchRun)
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.GET("/v1/apps/:app_id/config/:app_config_id", s.GetRunnerAppConfig)
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
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		cfg:             params.Cfg,
		v:               params.V,
		db:              params.DB,
		mw:              params.MW,
		l:               params.L,
		vcsHelpers:      params.VcsHelpers,
		helpers:         params.Helpers,
		installsHelpers: params.InstallsHelpers,
		accountsHelpers: params.AccountsHelpers,
		featuresClient:  params.FeaturesClient,
		temporalClient:  params.TemporalClient,
		queueClient:     params.QueueClient,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
