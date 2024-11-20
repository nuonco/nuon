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
	// manage apps
	// api.POST("/v1/apps/:app_id/action-workflows", s.CreateAppActionWorkflow)
	// api.GET("/v1/apps/:app_id/action-workflows", s.GetAppActionWorkflow)

	// work with actions directly
	// api.PATCH("/v1/actions/workflows/:workflow_id", s.UpdateActionWorkflow)
	// api.GET("/v1/actions/workflows/:workflow_id", s.GetAction)
	// api.DELETE("/v1/actions/:action_id", s.DeleteAction)

	// config versions
	// api.POST("/v1/actions/workflows/:action_workflow_id/configs", s.CreateActionConfig)
	// api.GET("/v1/actions/workflows/:action_workflow_id/configs/:config_id", s.GetActionConfig)
	// api.GET("/v1/actions/workflows/:action_workflow_id/configs", s.GetActionConfigs)

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
