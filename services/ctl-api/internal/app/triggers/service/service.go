package service

import (
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
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
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
	db           *gorm.DB
	cfg          *internal.Config
	l            *zap.Logger
	appsHelpers  *appshelpers.Helpers
	queueClient  *queueclient.Client
	httpClient   *http.Client
	snsVerifier  *eventsns.Verifier
	jwtMu        sync.Mutex
	jwtProviders map[string]*jwks.CachingProvider
}

var _ api.Service = (*service)(nil)

func New(p Params) *service {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	return &service{db: p.DB, cfg: p.Cfg, l: p.L, appsHelpers: p.AppsHelpers, queueClient: p.QueueClient, httpClient: httpClient, snsVerifier: eventsns.NewVerifier(httpClient), jwtProviders: make(map[string]*jwks.CachingProvider)}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	api.POST("/v1/event-ingress/:ingress_key", s.IngestEvent)

	triggers := require.Routes(api.Group("/v1/triggers"), permissions.APICtl)
	read := triggers.For(require.Organization(permissions.TriggerRead))
	credentials := triggers.For(require.Organization(permissions.TriggerCredentials))
	update := triggers.For(require.Organization(permissions.TriggerUpdate))
	triggers.POST("", require.Organization(permissions.TriggerCreate), s.CreateTrigger)
	read.GET("", s.ListTriggers)
	read.GET("/:trigger_id", s.GetTrigger)
	read.GET("/:trigger_id/events", s.ListTriggerEvents)
	read.GET("/:trigger_id/event-types", s.ListTriggerEventTypes)
	read.GET("/:trigger_id/rules", s.ListTriggerRules)
	read.GET("/:trigger_id/rules/:rule_id", s.GetTriggerRule)
	triggers.DELETE("/:trigger_id", require.Organization(permissions.TriggerDelete), s.DeleteTrigger)
	credentials.PATCH("/:trigger_id/ingress-url", s.GetTriggerIngressURL)
	credentials.POST("/:trigger_id/rotate-secret", s.RotateSecret)
	credentials.POST("/:trigger_id/rotate-ingress-url", s.RotateIngressURL)
	update.POST("/:trigger_id/enable", s.EnableTrigger)
	update.POST("/:trigger_id/disable", s.DisableTrigger)
	credentials.POST("/:trigger_id/secrets/:secret_id/revoke", s.RevokeSecret)
	credentials.PATCH("/:trigger_id/secrets/:secret_id/reveal", s.RevealSecret)

	read.GET("/events/:event_id", s.GetEvent)
	read.GET("/events/:event_id/raw", s.GetEventRaw)
	triggers.POST("/events/:event_id/replay", require.Organization(permissions.TriggerReplay), s.ReplayEvent)
	read.GET("/dispatches", s.ListDispatches)
	read.GET("/dispatches/:dispatch_id", s.GetDispatch)
	triggers.POST("/dispatches/:dispatch_id/retry", require.Point(permissions.TriggerRetry, s.dispatchOrgTarget), s.RetryDispatch)
	return nil
}

func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
