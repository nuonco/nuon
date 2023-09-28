package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	temporal "github.com/powertoolsdev/mono/pkg/temporal/client"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	mw             metrics.Writer
	temporalClient temporal.Client
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.POST("/v1/general/metrics", s.PublishMetrics)
	api.GET("/v1/general/current-user", s.GetCurrentUser)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.POST("/v1/general/provision-canary", s.ProvisionCanary)
	return nil
}

func New(v *validator.Validate, db *gorm.DB, mw metrics.Writer, l *zap.Logger, temporalClient temporal.Client) *service {
	return &service{
		l:              l,
		v:              v,
		mw:             mw,
		db:             db,
		temporalClient: temporalClient,
	}
}
