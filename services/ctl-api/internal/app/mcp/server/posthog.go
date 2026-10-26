package server

import (
	"context"
	"encoding/json"
	"strings"

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

func (s *Server) installIDMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		ctx = cctx.SetInstallIDContext(ctx, "")
		result, err := next(ctx, method, req)
		if method == "tools/call" {
			s.logMCPCapture(ctx, req)
		}
		return result, err
	}
}

func (s *Server) logMCPCapture(ctx context.Context, req mcp.Request) {
	if s == nil || s.l == nil || req == nil {
		return
	}
	agent, command := "", ""
	if extra := req.GetExtra(); extra != nil {
		agent = extra.Header.Get(agentclient.Header)
		command = extra.Header.Get(agentclient.CommandHeader)
	}
	cctx.GetLogger(ctx, s.l).Debug("mcp posthog capture",
		zap.String("tool", mcpToolName(req)),
		zap.String("install_id", cctx.InstallIDFromContext(ctx)),
		zap.String("agent", agent),
		zap.String("command", command),
	)
}

func attributionProperties(ctx context.Context, req *mcp.CallToolRequest, _ *mcp.CallToolResult, _ error) (posthog.Properties, error) {
	props := posthog.NewProperties()
	if req != nil && req.Extra != nil {
		if agent := req.Extra.Header.Get(agentclient.Header); agent != "" {
			props.Set("agent", agent)
		}
		if command := req.Extra.Header.Get(agentclient.CommandHeader); command != "" {
			props.Set("command", command)
		}
	}
	if id := mcpInstallID(ctx, req); id != "" {
		props.Set("install_id", id)
	}
	if len(props) == 0 {
		return nil, nil
	}
	return props, nil
}

// mcpInstallID prefers the ID the handler resolved. Tools that take install_id
// directly, without a lookup, fall back to that argument. A name in "install"
// is not an ID.
func mcpInstallID(ctx context.Context, req *mcp.CallToolRequest) string {
	if id := cctx.InstallIDFromContext(ctx); id != "" {
		return id
	}
	if req == nil || req.Params == nil || len(req.Params.Arguments) == 0 {
		return ""
	}
	// These tools take an install ID and do not resolve a name.
	switch req.Params.Name {
	case "list_deploys", "list_workflows":
	default:
		return ""
	}
	var args struct {
		InstallID string `json:"install_id"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return ""
	}
	return strings.TrimSpace(args.InstallID)
}
