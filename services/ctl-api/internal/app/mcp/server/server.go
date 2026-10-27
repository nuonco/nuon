package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	posthog "github.com/posthog/posthog-go"
	"github.com/posthog/posthog-go/posthogmcpsdk"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/agentclient"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/mcp/skills"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	controlplanemetrics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/productanalytics"
)

type Params struct {
	fx.In

	LC          fx.Lifecycle
	Shutdowner  fx.Shutdowner
	DB          *gorm.DB `name:"psql"`
	L           *zap.Logger
	Cfg         *internal.Config
	MW          metrics.Writer
	HTTPMetrics *controlplanemetrics.HTTPMetrics
	Analytics   *productanalytics.Client
	Services    []api.Service `group:"services"`
}

type NuonctlParams struct {
	fx.In

	LC          fx.Lifecycle
	Shutdowner  fx.Shutdowner
	DB          *gorm.DB `name:"psql"`
	L           *zap.Logger
	Cfg         *internal.Config
	MW          metrics.Writer
	HTTPMetrics *controlplanemetrics.HTTPMetrics
	Services    []api.MCPService `group:"nuonctl_mcp_services"`
}

type Server struct {
	db          *gorm.DB
	l           *zap.Logger
	cfg         *internal.Config
	mw          metrics.Writer
	httpMetrics *controlplanemetrics.HTTPMetrics
	mcpServices []api.MCPService
	httpServer  *http.Server
	schemaCache *mcp.SchemaCache
	posthog     posthog.Client

	posthogMiddleware *posthogmcpsdk.Middleware

	implementationName string
	requireEmployee    bool
	serverPurpose      string
	orgInstructions    string
}

func New(params Params) *Server {
	mcpServices := make([]api.MCPService, 0, len(params.Services))
	for _, svc := range params.Services {
		if mcpSvc, ok := svc.(api.MCPService); ok {
			mcpServices = append(mcpServices, mcpSvc)
		}
	}

	return newServer(
		params.LC,
		params.Shutdowner,
		params.DB,
		params.L,
		params.Cfg,
		params.MW,
		params.HTTPMetrics,
		params.Analytics,
		mcpServices,
		params.Cfg.MCPHTTPPort,
		"nuon-ctl",
		false,
		"Nuon control plane MCP server.",
		"Org-scoped calls require the X-Nuon-Org-ID header on that request. whoami and list_orgs do not. There is no stored org.",
	)
}

func NewNuonctl(params NuonctlParams) *Server {
	return newServer(
		params.LC,
		params.Shutdowner,
		params.DB,
		params.L,
		params.Cfg,
		params.MW,
		params.HTTPMetrics,
		nil,
		params.Services,
		params.Cfg.NuonctlMCPHTTPPort,
		"nuonctl",
		true,
		"Nuon employee-only nuonctl MCP server.",
		"Org-scoped calls require the X-Nuon-Org-ID header on that request. Run nuon orgs select so the CLI sends it, or set the header on each request.",
	)
}

func newServer(
	lc fx.Lifecycle,
	shutdowner fx.Shutdowner,
	db *gorm.DB,
	l *zap.Logger,
	cfg *internal.Config,
	mw metrics.Writer,
	httpMetrics *controlplanemetrics.HTTPMetrics,
	analyticsClient *productanalytics.Client,
	mcpServices []api.MCPService,
	port string,
	implementationName string,
	requireEmployee bool,
	serverPurpose string,
	orgInstructions string,
) *Server {
	s := &Server{
		db:                 db,
		l:                  l.Named("mcp"),
		cfg:                cfg,
		mw:                 mw,
		httpMetrics:        httpMetrics,
		mcpServices:        mcpServices,
		schemaCache:        mcp.NewSchemaCache(),
		implementationName: implementationName,
		requireEmployee:    requireEmployee,
		serverPurpose:      serverPurpose,
		orgInstructions:    orgInstructions,
	}

	if ph := analyticsClient.PostHog(); ph != nil && !requireEmployee {
		s.posthog = ph
		s.posthogMiddleware = s.newPostHogMiddleware()
	}

	mcpHandler := s.newMCPHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", s.healthHandler)
	mux.HandleFunc("/readyz", s.healthHandler)
	mux.HandleFunc("/.well-known/oauth-protected-resource", s.protectedResourceMetadataHandler)
	mux.Handle("/", s.authContextMiddleware(mcpHandler))

	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", port),
		Handler: s.otelMetricsMiddleware(s.metricsMiddleware(mux)),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.l.Info("starting MCP server", zap.String("addr", s.httpServer.Addr))
			go func() {
				if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.l.Error("MCP server error", zap.Error(err))
					_ = shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.l.Info("stopping MCP server")
			return s.httpServer.Shutdown(ctx)
		},
	})

	return s
}

// newMCPHandler builds the streamable-HTTP handler. The server is stateless: it
// neither reads nor sets Mcp-Session-Id, and each request gets a temporary
// session. Everything a request needs is carried by its bearer token and
// X-Nuon-Org-ID, so any replica can serve any request.
func (s *Server) newMCPHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(
		s.getServerForRequest,
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			Logger:    slog.Default(),
		},
	)
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) getServerForRequest(r *http.Request) *mcp.Server {
	// The middleware has already authenticated the request and injected the
	// account + selected org into the context.
	accountID := keys.CreatedByIDFromContext(r.Context())
	if accountID == "" {
		return nil
	}
	orgID := keys.OrgIDFromContext(r.Context())

	server := mcp.NewServer(&mcp.Implementation{
		Name:    s.implementationName,
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		SchemaCache:  s.schemaCache,
		Instructions: fmt.Sprintf("%s Authenticated as account %s in org %q. %s %s %s %s %s %s", s.serverPurpose, accountID, orgID, s.orgInstructions, api.MCPTimeInstructions, api.MCPPoliciesInstructions, api.MCPWatchInstructions, api.MCPAppConfigInstructions, skills.Instructions),
	})
	server.AddReceivingMiddleware(s.receivingMetricsMiddleware, s.loggingMiddleware)
	s.instrumentPostHog(server)

	for _, svc := range s.mcpServices {
		svc.RegisterMCPTools(server)
	}
	skills.Register(server)

	return server
}

// resolveOrg returns the request org from X-Nuon-Org-ID when the account can
// access it. An empty or inaccessible header leaves the org empty. Nothing is
// stored on the token.
func (s *Server) resolveOrg(acct *app.Account, headerOrg string) string {
	if headerOrg != "" && accountHasOrgAccess(acct, headerOrg) {
		return headerOrg
	}
	return ""
}

const traceIDHeaderKey = "X-Nuon-Trace-ID"

func (s *Server) authContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		traceID := r.Header.Get(traceIDHeaderKey)
		if traceID == "" {
			traceID = uuid.Must(uuid.NewV7()).String()
		}
		ctx = cctx.SetTraceIDContext(ctx, traceID)
		ctx = cctx.SetLoggerFields(ctx, []zap.Field{zap.Bool("mcp", true)})
		l := cctx.GetLogger(ctx, s.l)

		// Every request must carry a valid access token. A token failure returns
		// 401 + WWW-Authenticate so the client can discover the authorization
		// server and start the OAuth flow.
		acct, tok, err := s.authenticateToken(r)
		if err != nil {
			l.Warn("MCP auth failed", zap.Error(err))
			s.writeUnauthorized(w, r)
			return
		}
		if !s.accountAllowed(acct) {
			l.Warn("MCP employee access denied", zap.String("account_id", acct.ID))
			s.writeForbidden(w)
			return
		}

		orgID := s.resolveOrg(acct, r.Header.Get("X-Nuon-Org-ID"))

		ctx = cctx.SetAccountContext(ctx, acct)
		ctx = context.WithValue(ctx, keys.OrgIDCtxKey, orgID)
		ctx = keys.WithTokenRole(ctx, tok.Role)
		ctx = cctx.SetTokenTypeContext(ctx, tok.TokenType)
		ctx = context.WithValue(ctx, mcpHTTPRequestKey{}, mcpHTTPRequest{
			method: r.Method,
			path:   r.URL.Path,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// mcpHTTPRequest is the HTTP request that carried an MCP call. The agent log
// is written later, once the JSON-RPC method is known, and still includes these.
type mcpHTTPRequest struct {
	method string
	path   string
}

type mcpHTTPRequestKey struct{}

// loggingMiddleware writes the public API's "nuon cli agent" line once the
// JSON-RPC method is known. command is the MCP operation, such as
// tools/call/whoami, matching the metric endpoint name.
func (s *Server) loggingMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		s.logAgentOperation(ctx, req, mcpRPCEndpoint(method, mcpToolName(req)))
		return next(ctx, method, req)
	}
}

func (s *Server) logAgentOperation(ctx context.Context, req mcp.Request, command string) {
	if s == nil || s.l == nil || req == nil || command == "" {
		return
	}
	extra := req.GetExtra()
	if extra == nil {
		return
	}
	agent := extra.Header.Get(agentclient.Header)
	if agent == "" {
		return
	}
	fields := []zap.Field{
		zap.String("agent", agent),
		zap.String("command", command),
	}
	if info, ok := ctx.Value(mcpHTTPRequestKey{}).(mcpHTTPRequest); ok {
		fields = append(fields,
			zap.String("method", info.method),
			zap.String("path", info.path),
		)
	}
	cctx.GetLogger(ctx, s.l).Info("nuon cli agent", fields...)
}
