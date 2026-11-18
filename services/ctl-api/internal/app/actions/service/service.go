package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/access"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	comphelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V               *validator.Validate
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	Cfg             *internal.Config
	AccountsHelpers *accountshelpers.Helpers
	AcctClient      *account.Client
	VcsHelpers      *vcshelpers.Helpers
	CompHelpers     *comphelpers.Helpers
	Helpers         *actionshelpers.Helpers
	AppsHelpers     *appshelpers.Helpers
	InstallHelpers  *installhelpers.Helpers
	EndpointAudit   *apiPkg.EndpointAudit
	QueueClient     *queueclient.Client
	Audit           *audit.Emitter
}

type service struct {
	apiPkg.RouteRegister
	v               *validator.Validate
	db              *gorm.DB
	chDB            *gorm.DB
	cfg             *internal.Config
	accountsHelpers *accountshelpers.Helpers
	acctClient      *account.Client
	vcsHelpers      *vcshelpers.Helpers
	actionsHelpers  *actionshelpers.Helpers
	appsHelpers     *appshelpers.Helpers
	compHelpers     *comphelpers.Helpers
	installHelpers  *installhelpers.Helpers
	queueClient     *queueclient.Client
	audit           *audit.Emitter
}

var _ apiPkg.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// apps
	apps := api.Group("/v1/apps/:app_id", require.AppInOrg(s.db))
	{
		actions := require.Routes(apps.Group("/actions"), permissions.APICtl)
		{
			actions.POST("", require.Point(permissions.ActionCreate, s.actionAppTarget), s.CreateAppAction)
			actions.GET("", s.actionListRule(permissions.ActionInspect, actionaccess.Actions), s.GetAppActions)
			actions.GET("/label-keys", require.Point(permissions.ActionInspect, s.actionAppTarget), s.GetActionLabelKeys)
			actions.GET("/:action_id", require.Point(permissions.ActionInspect, s.actionTarget), s.GetAppAction)
			actions.PATCH("/:action_id", require.Point(permissions.ActionUpdate, s.actionIDTarget), s.UpdateAppAction)
			actions.DELETE("/:action_id", require.Point(permissions.ActionDelete, s.actionTarget), s.DeleteAppAction)
			actions.POST("/:action_id/labels", require.Point(permissions.ActionUpdate, s.actionIDTarget), s.AddAppActionLabels)
			actions.DELETE("/:action_id/labels", require.Point(permissions.ActionUpdate, s.actionIDTarget), s.RemoveAppActionLabels)
			actions.POST("/:action_id/configs", require.Point(permissions.ActionConfigCreate, s.actionConfigWriteTarget), s.CreateAppActionConfig)
			actions.GET("/:action_id/configs", s.actionListRule(permissions.ActionInspect, actionaccess.Configs), s.GetAppActionConfigs)
			actions.GET("/:action_id/latest-config", require.Point(permissions.ActionInspect, s.actionLatestConfigTarget), s.GetAppActionLatestConfig)
			actions.GET("/configs/:action_config_id", require.Point(permissions.ActionInspect, s.actionConfigTarget), s.GetAppActionConfig)
		}
	}

	// installs
	installs := api.Group("/v1/installs/:install_id")
	{
		// install action runs
		actionRuns := require.Routes(installs.Group("/actions/runs"), permissions.APICtl)
		{
			actionRuns.POST("", require.Point(permissions.ActionExecute, s.actionExecutionTarget), s.CreateInstallActionRun)
			actionRuns.GET("", s.actionListRule(permissions.InstallActionInspect, actionaccess.Runs), s.GetInstallActionRuns)
			actionRuns.GET("/:run_id", require.Point(permissions.InstallActionInspect, s.actionRunTarget), s.GetInstallActionRun)
			actionRuns.DELETE("/:run_id/logs", require.Point(permissions.ActionRunLogsDelete, s.actionLogsTarget), s.DeleteInstallActionRunLogs)
			actionRuns.GET("/:run_id/steps/:step_id", require.Point(permissions.InstallActionInspect, s.actionStepTarget), s.GetInstallActionRunStep)
		}

		// install actions
		installActions := require.Routes(installs.Group("/actions"), permissions.APICtl)
		{
			installActions.POST("/adhoc-run", require.Point(permissions.ActionAdhocExecute, s.actionInstallTarget), s.CreateAdHocAction)
			installActions.GET("", s.actionListRule(permissions.InstallActionInspect, actionaccess.InstallActions), s.GetInstallActions)
			installActions.For(require.Point(permissions.InstallActionInspect, s.installedActionTarget), s.actionListRule(permissions.InstallActionInspect, actionaccess.Runs)).GET("/:action_id/recent-runs", s.GetInstallActionRecentRuns)
			installActions.GET("/latest-runs", s.actionListRule(permissions.InstallActionInspect, actionaccess.InstallActions), s.GetInstallActionsLatestRuns)
			installActions.GET("/:action_id", require.Point(permissions.InstallActionInspect, s.installedActionTarget), s.GetInstallAction)
		}
	}

	// Deprecated routes
	// work with actions apps path
	deprecatedApps := require.Routes(api.Group("/v1/apps/:app_id", require.AppInOrg(s.db)), permissions.APICtl)
	{
		s.POST(deprecatedApps.For(require.Point(permissions.ActionCreate, s.actionAppTarget)), "/action-workflows", s.CreateAppActionWorkflow, apiPkg.APIContextTypePublic, true)
		s.GET(deprecatedApps.For(s.actionListRule(permissions.ActionInspect, actionaccess.Actions)), "/action-workflows", s.GetAppActionWorkflows, apiPkg.APIContextTypePublic, true)
		s.GET(deprecatedApps.For(require.Point(permissions.ActionInspect, s.appActionTarget)), "/action-workflows/:action_workflow_id", s.GetAppActionWorkflow, apiPkg.APIContextTypePublic, true)
	}

	// work with actions directly
	actionWorkflows := require.Routes(api.Group("/v1/action-workflows"), permissions.APICtl)
	{
		s.PATCH(actionWorkflows.For(require.Point(permissions.ActionUpdate, s.actionIDTarget)), "/:action_workflow_id", s.UpdateActionWorkflow, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows.For(require.Point(permissions.ActionInspect, s.actionTarget)), "/:action_workflow_id", s.GetActionWorkflow, apiPkg.APIContextTypePublic, true)
		s.DELETE(actionWorkflows.For(require.Point(permissions.ActionDelete, s.actionTarget)), "/:action_workflow_id", s.DeleteActionWorkflow, apiPkg.APIContextTypePublic, true)

		// config versions
		s.POST(actionWorkflows.For(require.Point(permissions.ActionConfigCreate, s.actionConfigWriteTarget)), "/:action_workflow_id/configs", s.CreateActionWorkflowConfig, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows.For(s.actionListRule(permissions.ActionInspect, actionaccess.Configs)), "/:action_workflow_id/configs", s.GetActionWorkflowConfigs, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows.For(require.Point(permissions.ActionInspect, s.actionLatestConfigTarget)), "/:action_workflow_id/latest-config", s.GetActionWorkflowLatestConfig, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows.For(require.Point(permissions.ActionInspect, s.actionConfigTarget)), "/configs/:action_workflow_config_id", s.GetActionWorkflowConfig, apiPkg.APIContextTypePublic, true)
	}

	// install runs (deprecated)
	deprecatedInstalls := api.Group("/v1/installs/:install_id")
	{
		actionWorkflowRuns := require.Routes(deprecatedInstalls.Group("/action-workflows/runs"), permissions.APICtl)
		{
			s.POST(actionWorkflowRuns.For(require.Point(permissions.ActionExecute, s.actionExecutionTarget)), "", s.CreateInstallActionWorkflowRun, apiPkg.APIContextTypePublic, true)
			s.GET(actionWorkflowRuns.For(s.actionListRule(permissions.InstallActionInspect, actionaccess.Runs)), "", s.GetInstallActionWorkflowRuns, apiPkg.APIContextTypePublic, true)
			s.GET(actionWorkflowRuns.For(require.Point(permissions.InstallActionInspect, s.actionRunTarget)), "/:run_id", s.GetInstallActionWorkflowRun, apiPkg.APIContextTypePublic, true)
			s.GET(actionWorkflowRuns.For(require.Point(permissions.InstallActionInspect, s.actionStepTarget)), "/:run_id/steps/:step_id", s.GetInstallActionWorkflowRunStep, apiPkg.APIContextTypePublic, true)
		}

		// install action workflows (deprecated)
		installActionWorkflows := require.Routes(deprecatedInstalls.Group("/action-workflows"), permissions.APICtl)
		{
			s.GET(installActionWorkflows.For(s.actionListRule(permissions.InstallActionInspect, actionaccess.InstallActions)), "", s.GetInstallActionWorkflows, apiPkg.APIContextTypePublic, true)
			s.GET(installActionWorkflows.For(require.Point(permissions.InstallActionInspect, s.installedActionTarget), s.actionListRule(permissions.InstallActionInspect, actionaccess.Runs)), "/:action_workflow_id/recent-runs", s.GetInstallActionWorkflowRecentRuns, apiPkg.APIContextTypePublic, true)
			s.GET(installActionWorkflows.For(s.actionListRule(permissions.InstallActionInspect, actionaccess.InstallActions)), "/latest-runs", s.GetInstallActionWorkflowsLatestRuns, apiPkg.APIContextTypePublic, true)
			s.GET(installActionWorkflows.For(require.Point(permissions.InstallActionInspect, s.installedActionTarget)), "/:action_workflow_id", s.GetInstallActionWorkflow, apiPkg.APIContextTypePublic, true)
		}
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.POST("/v1/action-workflows/:action_workflow_id/admin-restart", s.RestartAction)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	// Job SA has install-scoped grants, not org-level read/update.
	actionWorkflows := require.Routes(require.OrgOrInstallRoute(api), permissions.APIRunner).Group("/v1/action-workflows")
	{
		actionWorkflows.GET("/:workflow_id/latest-config", require.Point(permissions.ExecutionConfigRead, s.runnerActionLatestConfigTarget), s.GetActionWorkflowLatestConfig)
		actionWorkflows.GET("/configs/:action_workflow_config_id", require.Point(permissions.ExecutionConfigRead, s.runnerActionConfigTarget), s.GetActionWorkflowConfig)
	}

	installs := require.Routes(require.ResourceRoute(api, permissions.KindInstall, "install_id"), permissions.APIRunner).Group("/v1/installs/:install_id")
	{
		installs.PUT("/action-workflow-runs/:workflow_run_id/steps/:step_id", require.Point(permissions.ActionRunReport, s.actionStepTarget), s.UpdateInstallActionWorkflowRunStep)
		installs.GET("/action-workflows/runs/:run_id", require.Point(permissions.InstallActionInspect, s.actionRunTarget), s.GetInstallActionWorkflowRun)
	}

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
		cfg:             params.Cfg,
		v:               params.V,
		db:              params.DB,
		chDB:            params.CHDB,
		accountsHelpers: params.AccountsHelpers,
		acctClient:      params.AcctClient,
		vcsHelpers:      params.VcsHelpers,
		actionsHelpers:  params.Helpers,
		appsHelpers:     params.AppsHelpers,
		compHelpers:     params.CompHelpers,

		installHelpers: params.InstallHelpers,
		queueClient:    params.QueueClient,
		audit:          params.Audit,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
