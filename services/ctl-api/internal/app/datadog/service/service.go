// Package service exposes the org-scoped, dashboard- and CLI-facing public
// API for the Datadog integration: managing per-org DD connections (with
// a Test action), managing per-connection event subscriptions, and
// creating / managing the one-click "Alert on failure" managed monitors.
//
// Unlike the Slack package, there is no DD-side listener — DD never
// initiates traffic into Nuon, so there is no OAuth callback, slash
// command, or events endpoint. All routes are user-authenticated, mounted
// under /v1/orgs/:org_id/datadog/*.
package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
)

type Params struct {
	fx.In

	V             *validator.Validate
	DB            *gorm.DB `name:"psql"`
	MW            metrics.Writer
	L             *zap.Logger
	Cfg           *internal.Config
	DDClient      *ddclient.Client
	EndpointAudit *api.EndpointAudit
}

type service struct {
	api.RouteRegister

	v        *validator.Validate
	db       *gorm.DB
	mw       metrics.Writer
	l        *zap.Logger
	cfg      *internal.Config
	ddClient *ddclient.Client
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		v:        params.V,
		db:       params.DB,
		mw:       params.MW,
		l:        params.L,
		cfg:      params.Cfg,
		ddClient: params.DDClient,
	}
}

// RegisterPublicRoutes exposes the org-scoped public API (API key + Org ID
// auth) consumed by the dashboard UI and CLI.
//
// Naming note: in the ctl-api routing model, "public" means "the externally
// reachable, end-user-authenticated API surface" — NOT unauthenticated.
func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	g := ge.Group("/v1/orgs/:org_id/datadog")
	{
		// Connections: a per-org binding to a DD tenant. N per org so a
		// vendor's own DD + each customer's DD coexist.
		g.POST("/connections", s.CreateConnection)
		g.GET("/connections", s.ListConnections)
		g.GET("/connections/:connection_id", s.GetConnection)
		g.PATCH("/connections/:connection_id", s.UpdateConnection)
		g.DELETE("/connections/:connection_id", s.DeleteConnection)
		g.POST("/connections/:connection_id/test", s.TestConnection)

		// Event subscriptions: per-connection routing rules. Shape
		// mirrors Slack channel subscriptions one-to-one — same
		// Match + Interests contract, plus DD-specific tag/override
		// fields.
		g.GET("/event-subscriptions", s.ListEventSubscriptions)
		g.POST("/event-subscriptions", s.CreateEventSubscription)
		g.PATCH("/event-subscriptions/:sub_id", s.UpdateEventSubscription)
		g.DELETE("/event-subscriptions/:sub_id", s.DeleteEventSubscription)

		// Managed monitors: the one-click "Alert in Datadog" surface.
		// No update endpoint — the (connection, target, preset) tuple
		// is unique and immutable; users delete + recreate to change
		// the preset.
		g.GET("/managed-monitors", s.ListManagedMonitors)
		g.POST("/managed-monitors", s.CreateManagedMonitor)
		g.DELETE("/managed-monitors/:monitor_id", s.DeleteManagedMonitor)
	}
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error         { return nil }
func (s *service) RegisterInternalRoutes(api *gin.Engine) error       { return nil }
func (s *service) RegisterAuthRoutes(api *gin.Engine) error           { return nil }
func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(api *gin.Engine) error          { return nil }
