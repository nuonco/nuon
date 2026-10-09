package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	posthog "github.com/posthog/posthog-go"
	"github.com/posthog/posthog-go/posthogmcp"
	"github.com/posthog/posthog-go/posthogmcpsdk"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func newPostHogClient(key, host string) (posthog.Client, error) {
	if key == "" {
		return nil, nil
	}
	return posthog.NewWithConfig(key, posthog.Config{
		Endpoint: host,
		BeforeSend: func(msg posthog.Message) posthog.Message {
			if c, ok := msg.(posthog.Capture); ok {
				delete(c.Properties, "$mcp_error_message")
				return c
			}
			return msg
		},
	})
}

func (s *Server) instrumentPostHog(server *mcp.Server) {
	if s.posthogMiddleware == nil {
		return
	}
	server.AddReceivingMiddleware(s.posthogMiddleware.Receiving)
	server.AddSendingMiddleware(s.posthogMiddleware.Sending)
}

func (s *Server) newPostHogMiddleware() *posthogmcpsdk.Middleware {
	if s.posthog == nil {
		return nil
	}

	m := posthogmcpsdk.NewMiddleware(posthogmcp.New(s.posthog, posthogmcp.WithExceptionAutocapture(false)),
		posthogmcpsdk.WithServerInfo(s.implementationName, "1.0.0"),
		posthogmcpsdk.WithContextParameter(false),
		posthogmcpsdk.WithCaptureModel(false),
		posthogmcpsdk.WithConversationID(false),
		posthogmcpsdk.WithCaptureParameters(false),
		posthogmcpsdk.WithCaptureResponses(false),
		posthogmcpsdk.WithIdentity(func(ctx context.Context, _ *mcp.CallToolRequest) (posthogmcpsdk.Identity, error) {
			acct, err := cctx.AccountFromContext(ctx)
			if err != nil {
				return posthogmcpsdk.Identity{}, nil
			}
			id := posthogmcpsdk.Identity{DistinctID: acct.Email}
			if id.DistinctID == "" {
				id.DistinctID = acct.Subject
			}
			if orgID := keys.OrgIDFromContext(ctx); orgID != "" {
				id.Groups = posthog.NewGroups().Set("organization", orgID)
			}
			return id, nil
		}),
		posthogmcpsdk.WithErrorHandler(func(_ context.Context, err error) {
			s.l.Warn("posthog mcp analytics", zap.Error(err))
		}),
	)
	return &m
}
