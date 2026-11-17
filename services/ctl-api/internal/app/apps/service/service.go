package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/metric"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
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
	BlobService     blobstore.Service
	MeterProvider   metric.MeterProvider `optional:"true"`
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
	blobSvc         blobstore.Service
	sourceFileSize  metric.Int64Histogram
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	geAccess := require.Routes(ge, permissions.APICtl)
	geAccess.GET("/v1/branches", require.Collection(permissions.AppBranchRead, s.branchResources), s.GetOrgBranches)

	// manage apps
	apps := ge.Group("/v1/apps")
	appsAccess := require.Routes(apps, permissions.APICtl)
	{
		appsAccess.POST("", require.Point(permissions.AppCreate, s.appOrgTarget), s.CreateApp)
		appsAccess.GET("", require.Collection(permissions.AppInspect, s.appResources), s.GetApps)

		appByID := require.Routes(apps.Group("/:app_id", require.AppInOrg(s.db)), permissions.APICtl)
		appByID.PATCH("", s.namedAppRule(permissions.AppUpdate), s.UpdateApp)
		appByID.GET("", require.Point(permissions.AppInspect, s.appTarget), s.GetApp)
		appByID.DELETE("", s.appRule(permissions.AppDelete), s.DeleteApp)
	}

	// app-specific routes
	app := ge.Group("/v1/apps/:app_id", require.AppInOrg(s.db))
	// app configs
	appAccess := require.Routes(app, permissions.APICtl)
	{
		appAccess.GET("/template-config", s.namedAppRule(permissions.AppConfigRead), s.GetAppConfigTemplate)

		// deprecated singular route
		appConfig := require.Routes(app.Group("/config"), permissions.APICtl)
		{
			s.POST(appConfig.For(s.appRule(permissions.AppConfigWrite)), "", s.CreateAppConfig, api.APIContextTypePublic, true)                                           // deprecated singular route
			s.GET(appConfig.For(s.appRule(permissions.AppConfigRead)), "/:app_config_id", s.GetAppConfig, api.APIContextTypePublic, true)                                 // deprecated singular route
			s.PATCH(appConfig.For(s.appRule(permissions.AppAutomationManage)), "/:app_config_id", s.UpdateAppConfig, api.APIContextTypePublic, true)                      // deprecated singular route
			s.POST(appConfig.For(require.All(s.configInstallRequirements)), "/:app_config_id/update-installs", s.UpdateAppConfigInstalls, api.APIContextTypePublic, true) // deprecated singular route
			s.GET(appConfig.For(s.appRule(permissions.AppConfigRead)), "/:app_config_id/graph", s.GetAppConfigGraph, api.APIContextTypePublic, true)                      // deprecated singular route
		}

		appConfigs := require.Routes(app.Group("/configs"), permissions.APICtl)
		{
			appConfigs.GET("", s.appRule(permissions.AppConfigRead), s.GetAppConfigs)
			appConfigs.POST("", s.appRule(permissions.AppConfigWrite), s.CreateAppConfigV2)
			appConfigs.PATCH("/:config_id", s.appRule(permissions.AppAutomationManage), s.UpdateAppConfigV2)
			appConfigs.GET("/:config_id", s.appRule(permissions.AppConfigRead), s.GetAppConfigV2)
			appConfigs.POST("/:config_id/update-installs", require.All(s.configInstallRequirements), s.UpdateAppConfigInstallsV2)
			appConfigs.GET("/:config_id/graph", s.appRule(permissions.AppConfigRead), s.GetAppConfigGraphV2)
			appConfigs.POST("/:config_id/sync", s.appRule(permissions.AppAutomationManage), s.SyncAppConfig)
			appConfigs.POST("/:config_id/build", s.appRule(permissions.AppAutomationManage), s.BuildAppConfig)
			appConfigs.GET("/:config_id/diff", s.appRule(permissions.AppConfigRead), s.GetAppConfigDiff)
			appConfigs.GET("/:config_id/source-files/*path", s.appRule(permissions.AppConfigRead), s.GetAppConfigSourceFile)
		}

		// app sandbox builds
		sandboxBuilds := require.Routes(app.Group("/sandbox/builds"), permissions.APICtl)
		{
			sandboxBuilds.GET("", s.appRule(permissions.AppBuildRead), s.GetAppSandboxBuilds)
			sandboxBuilds.POST("", s.appRule(permissions.AppBuildCreate), s.CreateAppSandboxBuild)
			sandboxBuilds.GET("/:build_id", s.appRule(permissions.AppBuildRead), s.GetAppSandboxBuild)
		}

		// app sandbox management
		sandboxConfig := require.Routes(app.Group("/sandbox-config"), permissions.APICtl)
		{
			s.POST(sandboxConfig.For(s.childConfigRule()), "", s.CreateAppSandboxConfigV2, api.APIContextTypePublic, true) // deprecated singular route
		}

		sandboxConfigs := require.Routes(app.Group("/sandbox-configs"), permissions.APICtl)
		{
			sandboxConfigs.GET("", s.namedAppRule(permissions.AppConfigRead), s.GetAppSandboxConfigs)
			sandboxConfigs.POST("", s.childConfigRule(), s.CreateAppSandboxConfigV2)
		}
		appAccess.GET("/labels", s.namedAppRule(permissions.AppConfigRead), s.GetAppLabels)

		// app secrets configs management
		secretsConfigs := require.Routes(app.Group("/secrets-configs"), permissions.APICtl)
		{
			secretsConfigs.POST("", s.childConfigRule(), s.CreateAppSecretsConfig)
			secretsConfigs.GET("/:config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppSecretsConfig)
		}

		// app kubernetes contexts management
		kubernetesContextsConfigs := require.Routes(app.Group("/kubernetes-contexts-configs"), permissions.APICtl)
		{
			kubernetesContextsConfigs.POST("", s.childConfigRule(), s.CreateAppKubernetesContextsConfig)
		}

		// app stack configs
		stackConfigs := require.Routes(app.Group("/stack-configs"), permissions.APICtl)
		{
			stackConfigs.POST("", s.childConfigRule(), s.CreateAppStackConfig)
			stackConfigs.GET("/:config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppSecretsConfig)
		}

		// app policies management
		policiesConfigs := require.Routes(app.Group("/policies-configs"), permissions.APICtl)
		{
			policiesConfigs.GET("", s.namedAppRule(permissions.AppConfigRead), s.GetAppPoliciesConfigs)
			policiesConfigs.POST("", s.childConfigRule(), s.CreateAppPoliciesConfig)
			policiesConfigs.GET("/:config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppPoliciesConfig)
		}
		// individual policy config lookup
		appAccess.GET("/policy-config/:policy_config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppPolicyConfig)

		// app break glass
		breakGlassConfigs := require.Routes(app.Group("/break-glass-configs"), permissions.APICtl)
		{
			breakGlassConfigs.POST("", s.childConfigRule(), s.CreateAppBreakGlasssConfig)
			breakGlassConfigs.GET("/:config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppBreakGlassConfig)
		}

		// app permissions
		permissionsConfigs := require.Routes(app.Group("/permissions-configs"), permissions.APICtl)
		{
			permissionsConfigs.POST("", require.All(s.configApplyRequirements), s.CreateAppPermissionsConfig)
			permissionsConfigs.GET("/:config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppPermissionsConfig)
		}

		operationRoleConfigs := require.Routes(app.Group("/operation-role-configs"), permissions.APICtl)
		{
			operationRoleConfigs.POST("", s.childConfigRule(), s.CreateAppOperationRoleConfig)
			operationRoleConfigs.GET("/:operation_role_config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppOperationRoleConfigs)
		}

		// app runner management
		runnerConfigs := require.Routes(app.Group("/runner-configs"), permissions.APICtl)
		{
			runnerConfigs.POST("", require.All(s.configApplyRequirements), s.CreateAppRunnerConfig)
			runnerConfigs.GET("", s.namedAppRule(permissions.AppConfigRead), s.GetAppRunnerConfigs)
		}
		// app input management
		appAccess.POST("/input-config", s.childConfigRule(), s.CreateAppInputsConfig)

		inputConfigs := require.Routes(app.Group("/input-configs"), permissions.APICtl)
		{
			inputConfigs.GET("", s.namedAppRule(permissions.AppConfigRead), s.GetAppInputConfigs)
			inputConfigs.GET("/:input_config_id", s.namedAppRule(permissions.AppConfigRead), s.GetAppInputConfig)
		}

		// app secrets management
		s.POST(appAccess.For(s.namedAppRule(permissions.AppSecretsWrite)), "/secret", s.CreateAppSecretV2, api.APIContextTypePublic, true)
		// deprecated singular route
		appAccess.POST("/secrets", s.namedAppRule(permissions.AppSecretsWrite), s.CreateAppSecretV2)

		secret := require.Routes(app.Group("/secret"), permissions.APICtl)
		{
			s.DELETE(secret.For(s.namedAppRule(permissions.AppSecretsWrite)), "/:secret_id", s.DeleteAppSecret, api.APIContextTypePublic, true) // deprecated singular route
		}

		secrets := require.Routes(app.Group("/secrets"), permissions.APICtl)
		{
			secrets.GET("", s.namedAppRule(permissions.AppSecretsRead), s.GetAppSecrets)
			secrets.DELETE("/:secret_id", s.namedAppRule(permissions.AppSecretsWrite), s.DeleteAppSecretV2)

		}

		// app branches
		branches := require.Routes(app.Group("/branches"), permissions.APICtl)
		{
			branches.POST("", s.appRule(permissions.AppBranchWrite), s.CreateAppBranch)
			branches.GET("", s.appRule(permissions.AppBranchRead), s.GetAppBranches)
			branches.GET("/:app_branch_id", s.appRule(permissions.AppBranchRead), s.GetAppBranch)
			branches.PATCH("/:app_branch_id", s.appRule(permissions.AppBranchWrite), s.UpdateAppBranch)
			branches.DELETE("/:app_branch_id", require.All(s.branchDeleteRequirements), s.DeleteAppBranch)
			branches.GET("/:app_branch_id/configs", s.appRule(permissions.AppBranchRead), s.GetAppBranchAppConfigs)
			branches.POST("/:app_branch_id/configs", s.appRule(permissions.AppAutomationManage), s.CreateAppBranchConfig)
			branches.PATCH("/:app_branch_id/configs/:config_id", s.appRule(permissions.AppAutomationManage), s.UpdateAppBranchConfig)
			branches.GET("/:app_branch_id/latest-config", s.appRule(permissions.AppBranchRead), s.GetAppBranchLatestConfig)
			branches.GET("/:app_branch_id/preview-sources", s.appRule(permissions.AppBranchRead), s.GetAppBranchPreviewSources)
			branches.GET("/:app_branch_id/preview-install-candidates", s.appRule(permissions.AppBranchRead), s.GetAppBranchPreviewInstallCandidates)
			branches.POST("/:app_branch_id/runs", s.appRule(permissions.AppAutomationManage), s.TriggerAppBranchRun)
			branches.GET("/:app_branch_id/runs", s.appRule(permissions.AppBranchRead), s.GetAppBranchRuns)
			branches.GET("/:app_branch_id/runs/:run_id", s.appRule(permissions.AppBranchRead), s.GetAppBranchRun)
			branches.GET("/:app_branch_id/runs/:run_id/builds", s.appRule(permissions.AppBranchRead), s.GetAppBranchRunBuilds)
			branches.GET("/:app_branch_id/runs/:run_id/comparison", s.appRule(permissions.AppBranchRead), s.GetAppBranchRunComparison)
			branches.GET("/:app_branch_id/runs/:run_id/install-groups", s.appRule(permissions.AppBranchRead), s.GetAppBranchRunInstallGroups)
			branches.GET("/:app_branch_id/runs/:run_id/install-group-runs", s.appRule(permissions.AppBranchRead), s.GetInstallGroupRuns)
			branches.GET("/:app_branch_id/runs/:run_id/install-group-runs/:install_group_run_id", require.Point(permissions.AppBranchRead, s.installGroupRunTarget), s.GetInstallGroupRun)
			branches.POST("/:app_branch_id/sync-install-configs", require.All(s.branchConfigSyncRequirements), s.TriggerInstallConfigSync)
		}

		installSyncs := require.Routes(app.Group("/install-syncs"), permissions.APICtl)
		{
			installSyncs.GET("", s.appRule(permissions.AppBranchRead), s.GetAppInstallSyncs)
			installSyncs.POST("", s.appRule(permissions.AppAutomationManage), s.TriggerAppInstallSync)
			installSyncs.GET("/:sync_id", s.appRule(permissions.AppBranchRead), s.GetAppInstallSync)
			installSyncs.POST("/:sync_id/approvals/:approval_id/response", s.appRule(permissions.AppAutomationManage), s.RespondInstallCreationApproval)
		}

		installsConfigs := require.Routes(app.Group("/installs-configs"), permissions.APICtl)
		{
			installsConfigs.GET("", s.appRule(permissions.AppConfigRead), s.GetAppInstallsConfig)
			installsConfigs.POST("", s.appRule(permissions.AppAutomationManage), s.CreateAppInstallsConfig)
			installsConfigs.DELETE("/:config_id", s.appRule(permissions.AppAutomationManage), s.DeleteAppInstallsConfig)
		}
		// TODO deprecate - latest config routes
		appAccess.GET("/latest-break-glass-config", s.namedAppRule(permissions.AppConfigRead), s.GetLatestAppBreakGlassConfig)
		appAccess.GET("/runner-latest-config", s.namedAppRule(permissions.AppConfigRead), s.GetAppRunnerLatestConfig)
		appAccess.GET("/input-latest-config", s.namedAppRule(permissions.AppConfigRead), s.GetAppInputLatestConfig)
		appAccess.GET("/install-input-config", s.namedAppRule(permissions.AppConfigRead), s.GetAppInstallInputConfig)
		appAccess.GET("/latest-policies-config", s.namedAppRule(permissions.AppConfigRead), s.GetLatestAppPoliciesConfig)
		appAccess.GET("/latest-config", s.namedAppRule(permissions.AppConfigRead), s.GetAppLatestConfig)
		appAccess.GET("/sandbox-latest-config", s.namedAppRule(permissions.AppConfigRead), s.GetAppSandboxLatestConfig)
		appAccess.GET("/latest-secrets-config", s.namedAppRule(permissions.AppConfigRead), s.GetLatestAppSecretsConfig)
		appAccess.GET("/latest-permissions-config", s.namedAppRule(permissions.AppConfigRead), s.GetLatestAppPermissionsConfig)
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
	require.OrgOrInstallRoute(api).GET("/v1/apps/:app_id/config/:app_config_id", s.GetRunnerAppConfig)
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	var sourceFileSize metric.Int64Histogram
	if params.MeterProvider != nil {
		sourceFileSize, _ = params.MeterProvider.Meter("github.com/nuonco/nuon/services/ctl-api/internal/app/apps/service").Int64Histogram(
			"nuon.apps.config_source_file.size",
			metric.WithUnit("By"),
			metric.WithDescription("Size of app config source archive files served."),
			metric.WithExplicitBucketBoundaries(0, 128, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864),
		)
	}
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
		blobSvc:         params.BlobService,
		sourceFileSize:  sourceFileSize,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
