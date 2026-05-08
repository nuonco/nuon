package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	comphelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V              *validator.Validate
	DB             *gorm.DB `name:"psql"`
	Cfg            *internal.Config
	VcsHelpers     *vcshelpers.Helpers
	CompHelpers    *comphelpers.Helpers
	Helpers        *actionshelpers.Helpers
	EvClient       eventloop.Client
	InstallHelpers *installhelpers.Helpers
	EndpointAudit  *apiPkg.EndpointAudit
	FeaturesClient *features.Features
	QueueClient    *queueclient.Client
}

type service struct {
	apiPkg.RouteRegister
	v              *validator.Validate
	db             *gorm.DB
	cfg            *internal.Config
	vcsHelpers     *vcshelpers.Helpers
	actionsHelpers *actionshelpers.Helpers
	compHelpers    *comphelpers.Helpers
	evClient       eventloop.Client
	installHelpers *installhelpers.Helpers
	featuresClient *features.Features
	queueClient    *queueclient.Client
}

var _ apiPkg.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// apps
	apps := api.Group("/v1/apps/:app_id")
	{
		actions := apps.Group("/actions")
		{
			actions.POST("", s.CreateAppAction)
			actions.GET("", s.GetAppActions)
			actions.GET("/label-keys", s.GetActionLabelKeys)
			actions.GET("/:action_id", s.GetAppAction)
			actions.PATCH("/:action_id", s.UpdateAppAction)
			actions.DELETE("/:action_id", s.DeleteAppAction)
			actions.POST("/:action_id/labels", s.AddAppActionLabels)
			actions.DELETE("/:action_id/labels", s.RemoveAppActionLabels)
			actions.POST("/:action_id/configs", s.CreateAppActionConfig)
			actions.GET("/:action_id/configs", s.GetAppActionConfigs)
			actions.GET("/:action_id/latest-config", s.GetAppActionLatestConfig)
			actions.GET("/configs/:action_config_id", s.GetAppActionConfig)
		}
	}

	// installs
	installs := api.Group("/v1/installs/:install_id")
	{
		// install action runs
		actionRuns := installs.Group("/actions/runs")
		{
			actionRuns.POST("", s.CreateInstallActionRun)
			actionRuns.GET("", s.GetInstallActionRuns)
			actionRuns.GET("/:run_id", s.GetInstallActionRun)
			actionRuns.GET("/:run_id/steps/:step_id", s.GetInstallActionRunStep)
		}

		// install actions
		installActions := installs.Group("/actions")
		{
			installActions.POST("/adhoc-run", s.CreateAdHocAction)
			installActions.GET("", s.GetInstallActions)
			installActions.GET("/:action_id/recent-runs", s.GetInstallActionRecentRuns)
			installActions.GET("/latest-runs", s.GetInstallActionsLatestRuns)
			installActions.GET("/:action_id", s.GetInstallAction)
		}
	}

	// Deprecated routes
	// work with actions apps path
	deprecatedApps := api.Group("/v1/apps/:app_id")
	{
		s.POST(deprecatedApps, "/action-workflows", s.CreateAppActionWorkflow, apiPkg.APIContextTypePublic, true)
		s.GET(deprecatedApps, "/action-workflows", s.GetAppActionWorkflows, apiPkg.APIContextTypePublic, true)
		s.GET(deprecatedApps, "/action-workflows/:action_workflow_id", s.GetAppActionWorkflow, apiPkg.APIContextTypePublic, true)
	}

	// work with actions directly
	actionWorkflows := api.Group("/v1/action-workflows")
	{
		s.PATCH(actionWorkflows, "/:action_workflow_id", s.UpdateActionWorkflow, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows, "/:action_workflow_id", s.GetActionWorkflow, apiPkg.APIContextTypePublic, true)
		s.DELETE(actionWorkflows, "/:action_workflow_id", s.DeleteActionWorkflow, apiPkg.APIContextTypePublic, true)

		// config versions
		s.POST(actionWorkflows, "/:action_workflow_id/configs", s.CreateActionWorkflowConfig, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows, "/:action_workflow_id/configs", s.GetActionWorkflowConfigs, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows, "/:action_workflow_id/latest-config", s.GetActionWorkflowLatestConfig, apiPkg.APIContextTypePublic, true)
		s.GET(actionWorkflows, "/configs/:action_workflow_config_id", s.GetActionWorkflowConfig, apiPkg.APIContextTypePublic, true)
	}

	// install runs (deprecated)
	deprecatedInstalls := api.Group("/v1/installs/:install_id")
	{
		actionWorkflowRuns := deprecatedInstalls.Group("/action-workflows/runs")
		{
			s.POST(actionWorkflowRuns, "", s.CreateInstallActionWorkflowRun, apiPkg.APIContextTypePublic, true)
			s.GET(actionWorkflowRuns, "", s.GetInstallActionWorkflowRuns, apiPkg.APIContextTypePublic, true)
			s.GET(actionWorkflowRuns, "/:run_id", s.GetInstallActionWorkflowRun, apiPkg.APIContextTypePublic, true)
			s.GET(actionWorkflowRuns, "/:run_id/steps/:step_id", s.GetInstallActionWorkflowRunStep, apiPkg.APIContextTypePublic, true)
		}

		// install action workflows (deprecated)
		installActionWorkflows := deprecatedInstalls.Group("/action-workflows")
		{
			s.GET(installActionWorkflows, "", s.GetInstallActionWorkflows, apiPkg.APIContextTypePublic, true)
			s.GET(installActionWorkflows, "/:action_workflow_id/recent-runs", s.GetInstallActionWorkflowRecentRuns, apiPkg.APIContextTypePublic, true)
			s.GET(installActionWorkflows, "/latest-runs", s.GetInstallActionWorkflowsLatestRuns, apiPkg.APIContextTypePublic, true)
			s.GET(installActionWorkflows, "/:action_workflow_id", s.GetInstallActionWorkflow, apiPkg.APIContextTypePublic, true)
		}
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.POST("/v1/action-workflows/:action_workflow_id/admin-restart", s.RestartAction)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	// action workflows
	actionWorkflows := api.Group("/v1/action-workflows")
	{
		actionWorkflows.GET("/:workflow_id/latest-config", s.GetActionWorkflowLatestConfig)
		actionWorkflows.GET("/configs/:action_workflow_config_id", s.GetActionWorkflowConfig)
	}

	// installs
	installs := api.Group("/v1/installs/:install_id")
	{
		installs.PUT("/action-workflow-runs/:workflow_run_id/steps/:step_id", s.UpdateInstallActionWorkflowRunStep)
		installs.GET("/action-workflows/runs/:run_id", s.GetInstallActionWorkflowRun)
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
		cfg:            params.Cfg,
		v:              params.V,
		db:             params.DB,
		vcsHelpers:     params.VcsHelpers,
		actionsHelpers: params.Helpers,
		compHelpers:    params.CompHelpers,
		evClient:       params.EvClient,
		installHelpers: params.InstallHelpers,
		featuresClient: params.FeaturesClient,
		queueClient:    params.QueueClient,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
