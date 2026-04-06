package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	L        *zap.Logger
	DB       *gorm.DB `name:"psql"`
	V        *validator.Validate
	Helpers  *helpers.Helpers
	GhClient helpers.GithubClient `optional:"true"`
}

type service struct {
	l        *zap.Logger
	db       *gorm.DB
	v        *validator.Validate
	helpers  *helpers.Helpers
	ghClient helpers.GithubClient
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// vcs connections
	vcs := api.Group("/v1/vcs")
	{
		vcs.POST("/connection-callback", s.CreateConnectionCallback)

		// Webhook event receiver (public, no auth required)
		vcs.POST("/:vcs_connection_id/events", s.WriteEvent)

		connections := vcs.Group("/connections")
		{
			connections.POST("", s.CreateConnection)
			connections.GET("", s.GetConnections)
			connections.GET("/:connection_id", s.GetConnection)
			connections.GET("/:connection_id/branches", s.GetVCSConnectionRepoBranches)
			connections.DELETE("/:connection_id", s.DeleteConnection)
			connections.GET("/:connection_id/check-status", s.CheckConnectionStatus)
			connections.GET("/:connection_id/repos", s.ListConnectionRepos)
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

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	var ghClient helpers.GithubClient
	if params.GhClient != nil {
		ghClient = params.GhClient
	} else {
		ghClient = params.Helpers
	}

	return &service{
		v:        params.V,
		l:        params.L,
		db:       params.DB,
		helpers:  params.Helpers,
		ghClient: ghClient,
	}
}
