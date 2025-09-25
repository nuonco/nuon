package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	accountshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/accounts/helpers"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
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
	AccountsHelpers *accountshelpers.Helpers
	EvClient        eventloop.Client
}

type service struct {
	v               *validator.Validate
	db              *gorm.DB
	mw              metrics.Writer
	l               *zap.Logger
	cfg             *internal.Config
	vcsHelpers      *vcshelpers.Helpers
	helpers         *appshelpers.Helpers
	accountsHelpers *accountshelpers.Helpers
	evClient        eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
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
	api.PATCH("/v1/apps/:app_id/config/:app_config_id", s.UpdateAppConfig)
	api.POST("/v1/apps/:app_id/config/:app_config_id/update-installs", s.UpdateAppConfigInstalls)
	api.GET("/v1/apps/:app_id/config/:app_config_id/graph", s.GetAppConfigGraph)

	// app sandbox management
	api.POST("/v1/apps/:app_id/sandbox-config", s.CreateAppSandboxConfig)
	api.GET("/v1/apps/:app_id/sandbox-configs", s.GetAppSandboxConfigs)

	// app secrets management
	api.POST("/v1/apps/:app_id/secrets-configs", s.CreateAppSecretsConfig)
	api.GET("/v1/apps/:app_id/secrets-configs", s.GetAppSecretsConfig)

	// app stack configs
	api.POST("/v1/apps/:app_id/stack-configs", s.CreateAppStackConfig)
	api.GET("/v1/apps/:app_id/stack-configs/:config_id", s.GetAppSecretsConfig)

	// app policies management
	api.POST("/v1/apps/:app_id/policies-configs", s.CreateAppPoliciesConfig)
	api.GET("/v1/apps/:app_id/policies-configs/:app_policies_config_id", s.GetAppPoliciesConfig)

	// app break glass
	api.POST("/v1/apps/:app_id/break-glass-configs", s.CreateAppBreakGlasssConfig)
	api.GET("/v1/apps/:app_id/break-glass-configs/:app_break_glass_config_id", s.GetAppBreakGlassConfig)

	// app permissions
	api.POST("/v1/apps/:app_id/permissions-configs", s.CreateAppPermissionsConfig)
	api.GET("/v1/apps/:app_id/permissions-configs/:app_permissions_config_id", s.GetAppPermissionsConfig)

	// app runner management
	api.POST("/v1/apps/:app_id/runner-configs", s.CreateAppRunnerConfig)
	api.GET("/v1/apps/:app_id/runner-configs", s.GetAppRunnerConfigs)

	// app input management
	api.POST("/v1/apps/:app_id/input-config", s.CreateAppInputsConfig)
	api.GET("/v1/apps/:app_id/input-configs", s.GetAppInputConfigs)
	api.GET("/v1/apps/:app_id/input-configs/:input_config_id", s.GetAppInputConfig)

	// app secrets management
	api.POST("/v1/apps/:app_id/secret", s.CreateAppSecret)
	api.GET("/v1/apps/:app_id/secrets", s.GetAppSecrets)
	api.DELETE("/v1/apps/:app_id/secret/:secret_id", s.DeleteAppSecret)

	// app branches
	api.POST("/v1/apps/:app_id/branches", s.CreateAppBranch)
	api.GET("/v1/apps/:app_id/branches", s.GetAppBranches)
	api.GET("/v1/apps/:app_id/branches/:app_branch_id/configs", s.GetAppBranchAppConfigs)

	// TODO deprecate
	api.GET("/v1/apps/:app_id/latest-break-glass-config", s.GetLatestAppBreakGlassConfig)
	api.GET("/v1/apps/:app_id/runner-latest-config", s.GetAppRunnerLatestConfig)
	api.GET("/v1/apps/:app_id/input-latest-config", s.GetAppInputLatestConfig)
	api.GET("/v1/apps/:app_id/latest-policies-config", s.GetLatestAppPoliciesConfig)
	api.GET("/v1/apps/:app_id/latest-config", s.GetAppLatestConfig)
	api.GET("/v1/apps/:app_id/sandbox-latest-config", s.GetAppSandboxLatestConfig)
	api.GET("/v1/apps/:app_id/latest-secrets-config", s.GetLatestAppSecretsConfig)
	api.GET("/v1/apps/:app_id/latest-permissions-config", s.GetLatestAppPermissionsConfig)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/apps", s.GetAllApps)
	api.POST("/v1/apps/:app_id/admin-reprovision", s.AdminReprovisionApp)
	api.POST("/v1/apps/:app_id/admin-restart", s.RestartApp)
	api.POST("/v1/apps/:app_id/admin-config-graph", s.AdminConfigGraph)
	api.POST("/v1/app-branches/:app_branch_id/admin-test-app-branch-workflow", s.AdminTestAppBranchWorkflow)
	api.POST("/v1/app-branches/:app_branch_id/admin-restart", s.AdminRestartAppBranch)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.GET("/v1/apps/:app_id/config/:app_config_id", s.GetRunnerAppConfig)
	return nil
}

func New(params Params) *service {
	return &service{
		cfg:             params.Cfg,
		v:               params.V,
		db:              params.DB,
		mw:              params.MW,
		l:               params.L,
		vcsHelpers:      params.VcsHelpers,
		helpers:         params.Helpers,
		accountsHelpers: params.AccountsHelpers,
		evClient:        params.EvClient,
	}
}
