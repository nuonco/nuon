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
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/v1/vcs/:org_id/connections", s.GetOrgConnections)
	api.POST("/v1/vcs/:org_id/connection", s.CreateOrgConnection)
	api.GET("/v1/vcs/:org_id/connected-repos", s.GetOrgConnectedRepos)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func New(v *validator.Validate, db *gorm.DB, mw metrics.Writer, l *zap.Logger, ghClient *github.Client) *service {
	return &service{
		l:        l,
		db:       db,
		ghClient: ghClient,
	}
}
