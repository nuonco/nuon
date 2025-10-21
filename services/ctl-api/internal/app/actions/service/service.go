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
	api.POST("/v1/apps/:app_id/actions", s.CreateAppAction)
	api.GET("/v1/apps/:app_id/actions", s.GetAppActions)
	api.GET("/v1/apps/:app_id/actions/:action_id", s.GetAppAction)
	api.PATCH("/v1/apps/:app_id/actions/:action_id", s.UpdateAppAction)
	api.DELETE("/v1/apps/:app_id/actions/:action_id", s.DeleteAppAction)

	api.POST("/v1/apps/:app_id/actions/:action_id/configs", s.CreateAppActionConfig)
	api.GET("/v1/apps/:app_id/actions/:action_id/configs", s.GetAppActionConfigs)
	api.GET("/v1/apps/:app_id/actions/:action_id/latest-config", s.GetAppActionLatestConfig)
	api.GET("/v1/apps/:app_id/actions/configs/:action_config_id", s.GetAppActionConfig)

	// install runs
	api.POST("/v1/installs/:install_id/actions/runs", s.CreateInstallActionRun)
	api.GET("/v1/installs/:install_id/actions/runs", s.GetInstallActionRuns)
	api.GET("/v1/installs/:install_id/actions/runs/:run_id", s.GetInstallActionRun)
	api.GET("/v1/installs/:install_id/actions/runs/:run_id/steps/:step_id", s.GetInstallActionRunStep)

	// install actions
	api.GET("/v1/installs/:install_id/actions", s.GetInstallActions)
	api.GET("/v1/installs/:install_id/actions/:action_id/recent-runs", s.GetInstallActionRecentRuns)
	api.GET("/v1/installs/:install_id/actions/latest-runs", s.GetInstallActionsLatestRuns)
	api.POST("/v1/installs/:install_id/actions/:action_id", s.GetInstallAction)

	// Deprecated routes

	// work with actions apps path
	api.POST("/v1/apps/:app_id/action-workflows", s.CreateAppActionWorkflow)                 // deprecated
	api.GET("/v1/apps/:app_id/action-workflows", s.GetAppActionWorkflows)                    // deprecated
	api.GET("/v1/apps/:app_id/action-workflows/:action_workflow_id", s.GetAppActionWorkflow) // deprecated

	//  work with actions directly
	api.PATCH("/v1/action-workflows/:action_workflow_id", s.UpdateActionWorkflow)  // deprecated
	api.GET("/v1/action-workflows/:action_workflow_id", s.GetActionWorkflow)       // deprecated
	api.DELETE("/v1/action-workflows/:action_workflow_id", s.DeleteActionWorkflow) // deprecated

	// config versions
	api.POST("/v1/action-workflows/:action_workflow_id/configs", s.CreateActionWorkflowConfig)         // deprecated
	api.GET("/v1/action-workflows/:action_workflow_id/configs", s.GetActionWorkflowConfigs)            // deprecated
	api.GET("/v1/action-workflows/:action_workflow_id/latest-config", s.GetActionWorkflowLatestConfig) // deprecated
	api.GET("/v1/action-workflows/configs/:action_workflow_config_id", s.GetActionWorkflowConfig)      // deprecated

	// install runs
	api.POST("/v1/installs/:install_id/action-workflows/runs", s.CreateInstallActionWorkflowRun)                        // deprecated
	api.GET("/v1/installs/:install_id/action-workflows/runs", s.GetInstallActionWorkflowRuns)                           // deprecated
	api.GET("/v1/installs/:install_id/action-workflows/runs/:run_id", s.GetInstallActionWorkflowRun)                    // deprecated
	api.GET("/v1/installs/:install_id/action-workflows/runs/:run_id/steps/:step_id", s.GetInstallActionWorkflowRunStep) // deprecated

	// install action workflows
	api.GET("/v1/installs/:install_id/action-workflows", s.GetInstallActionWorkflows)                                          // deprecated
	api.GET("/v1/installs/:install_id/action-workflows/:action_workflow_id/recent-runs", s.GetInstallActionWorkflowRecentRuns) // deprecated
	api.GET("/v1/installs/:install_id/action-workflows/latest-runs", s.GetInstallActionWorkflowsLatestRuns)                    // deprecated
	api.POST("/v1/installs/:install_id/action-workflows/:action_workflow_id", s.GetInstallActionWorkflow)                      // deprecated

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.POST("/v1/action-workflows/:action_workflow_id/admin-restart", s.RestartAction)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.GET("/v1/action-workflows/:workflow_id/latest-config", s.GetActionWorkflowLatestConfig)
	api.GET("/v1/action-workflows/configs/:action_workflow_config_id", s.GetActionWorkflowConfig)

	api.PUT("/v1/installs/:install_id/action-workflow-runs/:workflow_run_id/steps/:step_id", s.UpdateInstallActionWorkflowRunStep)
	api.GET("/v1/installs/:install_id/action-workflows/runs/:run_id", s.GetInstallActionWorkflowRun)
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
