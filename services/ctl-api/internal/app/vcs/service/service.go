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
	api.POST("/v1/vcs/connection-callback", s.CreateConnectionCallback)
	api.POST("/v1/vcs/connections", s.CreateConnection)

	api.GET("/v1/vcs/connections", s.GetConnections)
	api.GET("/v1/vcs/connections/:connection_id", s.GetConnection)
	api.DELETE("/v1/vcs/connections/:connection_id", s.DeleteConnection)

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
