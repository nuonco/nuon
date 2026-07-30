package service

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	eventsns "github.com/nuonco/nuon/pkg/events/sns"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	Cfg         *internal.Config
	L           *zap.Logger
	AppsHelpers *appshelpers.Helpers
	QueueClient *queueclient.Client
	Features    *features.Features
}

type service struct {
	db           *gorm.DB
	cfg          *internal.Config
	l            *zap.Logger
	appsHelpers  *appshelpers.Helpers
	queueClient  *queueclient.Client
	features     *features.Features
	httpClient   *http.Client
	snsVerifier  *eventsns.Verifier
	jwtMu        sync.Mutex
	jwtProviders map[string]*jwks.CachingProvider
}

var _ api.Service = (*service)(nil)

func New(p Params) *service {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	return &service{db: p.DB, cfg: p.Cfg, l: p.L, appsHelpers: p.AppsHelpers, queueClient: p.QueueClient, features: p.Features, httpClient: httpClient, snsVerifier: eventsns.NewVerifier(httpClient), jwtProviders: make(map[string]*jwks.CachingProvider)}
}

func (s *service) requireTriggers(ctx *gin.Context) {
	enabled, err := s.features.FeatureEnabled(ctx, app.OrgFeatureTriggers)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}
	if !enabled {
		ctx.Error(errors.New("triggers feature is not enabled"))
		ctx.Abort()
		return
	}
	ctx.Next()
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	api.POST("/v1/event-ingress/:ingress_key", s.IngestEvent)

	triggers := api.Group("/v1/triggers")
	triggers.Use(s.requireTriggers)
	triggers.POST("", s.CreateTrigger)
	triggers.GET("", s.ListTriggers)
	triggers.GET("/:trigger_id", s.GetTrigger)
	triggers.GET("/:trigger_id/events", s.ListTriggerEvents)
	triggers.GET("/:trigger_id/event-types", s.ListTriggerEventTypes)
	triggers.GET("/:trigger_id/rules", s.ListTriggerRules)
	triggers.GET("/:trigger_id/rules/:rule_id", s.GetTriggerRule)
	triggers.DELETE("/:trigger_id", s.DeleteTrigger)
	triggers.PATCH("/:trigger_id/ingress-url", s.GetTriggerIngressURL)
	triggers.POST("/:trigger_id/rotate-secret", s.RotateSecret)
	triggers.POST("/:trigger_id/rotate-ingress-url", s.RotateIngressURL)
	triggers.POST("/:trigger_id/enable", s.EnableTrigger)
	triggers.POST("/:trigger_id/disable", s.DisableTrigger)
	triggers.POST("/:trigger_id/secrets/:secret_id/revoke", s.RevokeSecret)
	triggers.PATCH("/:trigger_id/secrets/:secret_id/reveal", s.RevealSecret)

	triggerRoutes := api.Group("/v1/triggers")
	triggerRoutes.Use(s.requireTriggers)
	triggerRoutes.GET("/events/:event_id", s.GetEvent)
	triggerRoutes.GET("/events/:event_id/raw", s.GetEventRaw)
	triggerRoutes.POST("/events/:event_id/replay", s.ReplayEvent)
	triggerRoutes.GET("/dispatches", s.ListDispatches)
	triggerRoutes.GET("/dispatches/:dispatch_id", s.GetDispatch)
	triggerRoutes.POST("/dispatches/:dispatch_id/retry", s.RetryDispatch)
	return nil
}

func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
