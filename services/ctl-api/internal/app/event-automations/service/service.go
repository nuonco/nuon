package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	Cfg         *internal.Config
	L           *zap.Logger
	AppsHelpers *appshelpers.Helpers
	QueueClient *queueclient.Client
}

type service struct {
	db          *gorm.DB
	cfg         *internal.Config
	l           *zap.Logger
	appsHelpers *appshelpers.Helpers
	queueClient *queueclient.Client
}

var _ api.Service = (*service)(nil)

func New(p Params) *service {
	return &service{db: p.DB, cfg: p.Cfg, l: p.L, appsHelpers: p.AppsHelpers, queueClient: p.QueueClient}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	api.POST("/v1/event-ingress/:ingress_key", s.IngestEvent)

	sources := api.Group("/v1/apps/:app_id/event-sources")
	sources.POST("", s.CreateEventSource)
	sources.GET("", s.ListEventSources)
	sources.GET("/:event_source_id", s.GetEventSource)
	sources.POST("/:event_source_id/rotate-secret", s.RotateSecret)
	sources.POST("/:event_source_id/enable", s.EnableEventSource)
	sources.POST("/:event_source_id/disable", s.DisableEventSource)
	sources.POST("/:event_source_id/secrets/:secret_id/revoke", s.RevokeSecret)
	return nil
}

func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
