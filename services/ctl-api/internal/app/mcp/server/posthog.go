package server

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	posthog "github.com/posthog/posthog-go"
	"github.com/posthog/posthog-go/posthogmcp"
	"github.com/posthog/posthog-go/posthogmcpsdk"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/agentclient"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

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
		posthogmcpsdk.WithProperties(attributionProperties),
		posthogmcpsdk.WithErrorHandler(func(_ context.Context, err error) {
			s.l.Warn("posthog mcp analytics", zap.Error(err))
		}),
	)
	return &m
}

func attributionProperties(_ context.Context, req *mcp.CallToolRequest, _ *mcp.CallToolResult, _ error) (posthog.Properties, error) {
	if req == nil || req.Extra == nil {
		return nil, nil
	}
	props := posthog.NewProperties()
	if agent := mcpAgent(req.Extra.Header); agent != "" {
		props.Set("agent", agent)
	}
	if command := req.Extra.Header.Get(agentclient.CommandHeader); command != "" {
		props.Set("command", command)
	}
	if len(props) == 0 {
		return nil, nil
	}
	return props, nil
}

func mcpAgent(h http.Header) string {
	if agent := agentclient.Normalize(h.Get(agentclient.Header)); agent != "" {
		return agent
	}
	if agent, ok := agentclient.FromUserAgent(h.Get("User-Agent")); ok {
		return agent
	}
	return ""
}
