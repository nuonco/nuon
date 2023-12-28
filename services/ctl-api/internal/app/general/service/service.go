package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	temporal "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	temporalClient temporal.Client
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.POST("/v1/general/metrics", s.PublishMetrics)
	api.GET("/v1/general/current-user", s.GetCurrentUser)
	api.GET("/v1/general/cli-config", s.GetCLIConfig)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.POST("/v1/general/provision-canary", s.ProvisionCanary)
	api.POST("/v1/general/deprovision-canary", s.DeprovisionCanary)
	api.POST("/v1/general/start-canary-cron", s.StartCanaryCron)
	api.POST("/v1/general/stop-canary-cron", s.StopCanaryCron)

	api.POST("/v1/general/integration-user", s.CreateIntegrationUser)
	api.POST("/v1/general/canary-user", s.CreateCanaryUser)

	return nil
}

func New(v *validator.Validate, db *gorm.DB, mw metrics.Writer, l *zap.Logger, temporalClient temporal.Client, cfg *internal.Config) *service {
	return &service{
		l:              l,
		v:              v,
		mw:             mw,
		db:             db,
		temporalClient: temporalClient,
		cfg:            cfg,
	}
}
