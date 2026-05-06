// Package service exposes the org-scoped, dashboard-facing public API for the
// Slack integration: enumerating installations linked to an org, listing /
// creating channel subscriptions, and kicking off the OAuth install flow.
//
// The Slack-side surface (OAuth callback, slash commands, Events API) is
// handled separately on the dedicated Slack listener via RegisterSlackRoutes
// and is intentionally out of scope here (Phase 4).
package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/signing"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/statejwt"
)

type Params struct {
	fx.In

	V             *validator.Validate
	DB            *gorm.DB `name:"psql"`
	MW            metrics.Writer
	L             *zap.Logger
	Cfg           *internal.Config
	SlackClient   *slackclient.Client
	StateJWT      *statejwt.Encoder
	EndpointAudit *api.EndpointAudit
}

type service struct {
	api.RouteRegister

	v           *validator.Validate
	db          *gorm.DB
	mw          metrics.Writer
	l           *zap.Logger
	cfg         *internal.Config
	slackClient *slackclient.Client
	stateJWT    *statejwt.Encoder
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		v:           params.V,
		db:          params.DB,
		mw:          params.MW,
		l:           params.L,
		cfg:         params.Cfg,
		slackClient: params.SlackClient,
		stateJWT:    params.StateJWT,
	}
}

// RegisterPublicRoutes exposes the org-scoped public API (API key + Org ID
// auth) consumed by the dashboard UI.
//
// Naming note: in the ctl-api routing model, "public" means "the externally
// reachable, end-user-authenticated API surface" (vs. the runner / internal /
// auth listeners). It is NOT unauthenticated.
func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	orgs := ge.Group("/v1/orgs/:org_id/slack")
	{
		orgs.GET("/install-url", s.GetInstallURL)
		orgs.GET("/installations", s.ListInstallations)
		orgs.GET("/installations/:installation_id/channels", s.ListChannels)

		orgs.GET("/org-links", s.ListOrgLinks)
		orgs.POST("/org-links", s.CreateOrgLink)
		orgs.DELETE("/org-links/:link_id", s.DeleteOrgLink)

		orgs.GET("/channel-subscriptions", s.ListChannelSubscriptions)
		orgs.POST("/channel-subscriptions", s.CreateChannelSubscription)
		orgs.PATCH("/channel-subscriptions/:sub_id", s.UpdateChannelSubscription)
		orgs.DELETE("/channel-subscriptions/:sub_id", s.DeleteChannelSubscription)
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error         { return nil }
func (s *service) RegisterInternalRoutes(api *gin.Engine) error       { return nil }
func (s *service) RegisterAuthRoutes(api *gin.Engine) error           { return nil }
func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error { return nil }

// RegisterSlackRoutes wires the Slack-facing listener: OAuth callback, slash
// commands, and Events API webhooks. This runs on the dedicated Slack HTTP
// server (cfg.SlackHTTPPort) and is exposed to the public internet.
//
// Auth model — distinct from /v1/*:
//
//   - The OAuth callback is NOT signed by Slack (browser redirect). Trust
//     comes from the state JWT signature verified inside the handler.
//   - Slash commands and events ARE signed by Slack. We mount
//     signing.Middleware(cfg.SlackSigningSecret) at the route-group level
//     (not as a global SlackMiddleware) so the OAuth callback is
//     unaffected. cfg.SlackMiddlewares should remain limited to
//     non-auth concerns (logging, cors, recovery) — see api/base.go.
func (s *service) RegisterSlackRoutes(ge *gin.Engine) error {
	ge.GET("/slack/oauth/callback", s.SlackOAuthCallback)

	signMW, err := signing.Middleware(s.cfg.SlackSigningSecret)
	if err != nil {
		return fmt.Errorf("slack signing middleware: %w", err)
	}
	signed := ge.Group("", signMW)
	signed.POST("/slack/commands/nuon", s.SlackSlashCommand)
	signed.POST("/slack/events", s.SlackEvents)
	// Single request_url for all interactive surfaces (modals, buttons,
	// select menus, dynamic options). Stub returns 200 today; modal
	// handlers land in subsequent commits within this PR.
	signed.POST("/slack/interactions", s.SlackInteractions)

	return nil
}
