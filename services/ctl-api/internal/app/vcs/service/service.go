package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	L       *zap.Logger
	DB      *gorm.DB `name:"psql"`
	V       *validator.Validate
	Helpers *helpers.Helpers
}

type service struct {
	l       *zap.Logger
	db      *gorm.DB
	v       *validator.Validate
	helpers *helpers.Helpers
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// vcs connections
	vcs := api.Group("/v1/vcs")
	{
		vcs.POST("/connection-callback", s.CreateConnectionCallback)

		connections := vcs.Group("/connections")
		{
			connections.POST("", s.CreateConnection)
			connections.GET("", s.GetConnections)
			connections.GET("/:connection_id", s.GetConnection)
			connections.DELETE("/:connection_id", s.DeleteConnection)
		}
	}
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		v:       params.V,
		l:       params.L,
		db:      params.DB,
		helpers: params.Helpers,
	}
}
