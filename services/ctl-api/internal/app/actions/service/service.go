package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	actionshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/helpers"
	comphelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	V           *validator.Validate
	DB          *gorm.DB `name:"psql"`
	Cfg         *internal.Config
	VcsHelpers  *vcshelpers.Helpers
	CompHelpers *comphelpers.Helpers
	Helpers     *actionshelpers.Helpers
	EvClient    eventloop.Client
}

type service struct {
	v              *validator.Validate
	db             *gorm.DB
	cfg            *internal.Config
	vcsHelpers     *vcshelpers.Helpers
	actionsHelpers *actionshelpers.Helpers
	compHelpers    *comphelpers.Helpers
	evClient       eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// apps
	apps := api.Group("/v1/apps/:app_id")
	{
		actions := apps.Group("/actions")
		{
			actions.POST("", s.CreateAppAction)
			actions.GET("", s.GetAppActions)
			actions.GET("/:action_id", s.GetAppAction)
			actions.PATCH("/:action_id", s.UpdateAppAction)
			actions.DELETE("/:action_id", s.DeleteAppAction)
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
			installActions.GET("", s.GetInstallActions)
			installActions.GET("/:action_id/recent-runs", s.GetInstallActionRecentRuns)
			installActions.GET("/latest-runs", s.GetInstallActionsLatestRuns)
			installActions.POST("/:action_id", s.GetInstallAction)
		}
	}

	// Deprecated routes
	// work with actions apps path
	deprecatedApps := api.Group("/v1/apps/:app_id")
	{
		deprecatedApps.POST("/action-workflows", s.CreateAppActionWorkflow)                 // deprecated
		deprecatedApps.GET("/action-workflows", s.GetAppActionWorkflows)                    // deprecated
		deprecatedApps.GET("/action-workflows/:action_workflow_id", s.GetAppActionWorkflow) // deprecated
	}

	// work with actions directly
	actionWorkflows := api.Group("/v1/action-workflows")
	{
		actionWorkflows.PATCH("/:action_workflow_id", s.UpdateActionWorkflow)  // deprecated
		actionWorkflows.GET("/:action_workflow_id", s.GetActionWorkflow)       // deprecated
		actionWorkflows.DELETE("/:action_workflow_id", s.DeleteActionWorkflow) // deprecated

		// config versions
		actionWorkflows.POST("/:action_workflow_id/configs", s.CreateActionWorkflowConfig)         // deprecated
		actionWorkflows.GET("/:action_workflow_id/configs", s.GetActionWorkflowConfigs)            // deprecated
		actionWorkflows.GET("/:action_workflow_id/latest-config", s.GetActionWorkflowLatestConfig) // deprecated
		actionWorkflows.GET("/configs/:action_workflow_config_id", s.GetActionWorkflowConfig)      // deprecated
	}

	// install runs (deprecated)
	deprecatedInstalls := api.Group("/v1/installs/:install_id")
	{
		actionWorkflowRuns := deprecatedInstalls.Group("/action-workflows/runs")
		{
			actionWorkflowRuns.POST("", s.CreateInstallActionWorkflowRun)                        // deprecated
			actionWorkflowRuns.GET("", s.GetInstallActionWorkflowRuns)                           // deprecated
			actionWorkflowRuns.GET("/:run_id", s.GetInstallActionWorkflowRun)                    // deprecated
			actionWorkflowRuns.GET("/:run_id/steps/:step_id", s.GetInstallActionWorkflowRunStep) // deprecated
		}

		// install action workflows (deprecated)
		installActionWorkflows := deprecatedInstalls.Group("/action-workflows")
		{
			installActionWorkflows.GET("", s.GetInstallActionWorkflows)                                          // deprecated
			installActionWorkflows.GET("/:action_workflow_id/recent-runs", s.GetInstallActionWorkflowRecentRuns) // deprecated
			installActionWorkflows.GET("/latest-runs", s.GetInstallActionWorkflowsLatestRuns)                    // deprecated
			installActionWorkflows.POST("/:action_workflow_id", s.GetInstallActionWorkflow)                      // deprecated
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

func New(params Params) *service {
	return &service{
		cfg:            params.Cfg,
		v:              params.V,
		db:             params.DB,
		vcsHelpers:     params.VcsHelpers,
		actionsHelpers: params.Helpers,
		compHelpers:    params.CompHelpers,
		evClient:       params.EvClient,
	}
}
