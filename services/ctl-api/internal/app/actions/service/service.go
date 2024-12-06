package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	V          *validator.Validate
	DB         *gorm.DB `name:"psql"`
	Cfg        *internal.Config
	VcsHelpers *vcshelpers.Helpers
	EvClient   eventloop.Client
}

type service struct {
	v          *validator.Validate
	db         *gorm.DB
	cfg        *internal.Config
	vcsHelpers *vcshelpers.Helpers
	evClient   eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// work with actions apps path
	api.POST("/v1/apps/:app_id/action-workflows", s.CreateAppActionWorkflow)
	api.GET("/v1/apps/:app_id/action-workflows", s.GetAppActionWorkflows)

	//work with actions directly
	api.PATCH("/v1/action-workflows/:action_workflow_id", s.UpdateActionWorkflow)
	api.GET("/v1/action-workflows/:action_workflow_id", s.GetActionWorkflow)
	api.DELETE("/v1/action-workflows/:action_workflow_id", s.DeleteActionWorkflow)

	// config versions
	api.POST("/v1/action-workflows/:action_workflow_id/configs", s.CreateActionWorkflowConfig)
	api.GET("/v1/action-workflows/:action_workflow_id/configs", s.GetActionWorkflowConfigs)
	api.GET("/v1/action-workflows/configs/:action_workflow_config_id", s.GetActionWorkflowConfig)

	// install runs
	api.POST("/v1/installs/:install_id/action-workflows/runs", s.CreateInstallActionWorkflowRun)
	api.GET("/v1/installs/:install_id/action-workflows/runs", s.GetInstallActionWorkflowRuns)
	api.GET("/v1/installs/:install_id/action-workflows/runs/:run_id", s.GetInstallActionWorkflowRun)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// api.GET("/v1/actions", s.GetAllActions)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		cfg:        params.Cfg,
		v:          params.V,
		db:         params.DB,
		vcsHelpers: params.VcsHelpers,
		evClient:   params.EvClient,
	}
}
