package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	l        *zap.Logger
	db       *gorm.DB
	ghClient *github.Client
	v        *validator.Validate
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/v1/vcs/connections", s.GetConnections)
	api.GET("/v1/vcs/connections/:connection_id", s.GetConnection)
	api.GET("/v1/vcs/connected-repos", s.GetAllConnectedRepos)

	api.POST("/v1/vcs/connections", s.CreateConnection)
	api.POST("/v1/vcs/connection-callback", s.CreateConnectionCallback)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func New(v *validator.Validate, db *gorm.DB, mw metrics.Writer, l *zap.Logger, ghClient *github.Client) *service {
	return &service{
		v:        v,
		l:        l,
		db:       db,
		ghClient: ghClient,
	}
}
