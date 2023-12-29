package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	l       *zap.Logger
	db      *gorm.DB
	v       *validator.Validate
	helpers *helpers.Helpers
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

func New(v *validator.Validate,
	db *gorm.DB,
	mw metrics.Writer,
	l *zap.Logger,
	helpers *helpers.Helpers,
) *service {
	return &service{
		v:       v,
		l:       l,
		db:      db,
		helpers: helpers,
	}
}
