package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type Params struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner
	DB         *gorm.DB `name:"psql"`
	L          *zap.Logger
	Cfg        *internal.Config
	Services   []api.Service `group:"services"`
}

// mcpSessionTTL bounds how long an idle MCP session (its cached org selection)
// is retained. It matches the streamable-HTTP SessionTimeout.
const mcpSessionTTL = 30 * time.Minute

type Server struct {
	db         *gorm.DB
	l          *zap.Logger
	cfg        *internal.Config
	services   []api.Service
	httpServer *http.Server

	mu          sync.RWMutex
	sessions    map[string]*authResult
	stopJanitor chan struct{}
}

func New(params Params) *Server {
	s := &Server{
		db:          params.DB,
		l:           params.L.Named("mcp"),
		cfg:         params.Cfg,
		services:    params.Services,
		sessions:    make(map[string]*authResult),
		stopJanitor: make(chan struct{}),
	}

	mcpHandler := mcp.NewStreamableHTTPHandler(
		s.getServerForRequest,
		&mcp.StreamableHTTPOptions{
			SessionTimeout: mcpSessionTTL,
			Logger:         slog.Default(),
		},
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", s.healthHandler)
	mux.HandleFunc("/readyz", s.healthHandler)
	mux.HandleFunc("/.well-known/oauth-protected-resource", s.protectedResourceMetadataHandler)
	mux.Handle("/", s.authContextMiddleware(mcpHandler))

	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", params.Cfg.MCPHTTPPort),
		Handler: mux,
	}

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.l.Info("starting MCP server", zap.String("addr", s.httpServer.Addr))
			go func() {
				if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.l.Error("MCP server error", zap.Error(err))
					_ = params.Shutdowner.Shutdown()
				}
			}()
			go s.runSessionJanitor()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.l.Info("stopping MCP server")
			close(s.stopJanitor)
			return s.httpServer.Shutdown(ctx)
		},
	})

	return s
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
		Name:    "nuon-ctl",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Instructions: fmt.Sprintf("Nuon control plane MCP server. Authenticated as account %s in org %q. If no org is selected, call list_orgs then select_org.", accountID, orgID),
	})

	for _, svc := range s.services {
		if mcpSvc, ok := svc.(api.MCPService); ok {
			mcpSvc.RegisterMCPTools(server)
		}
	}

	return server
}

// resolveOrg determines the active org for a request: a previously selected org
// on the session wins; otherwise fall back to a valid X-Nuon-Org-ID header, or
// auto-select the account's only org. The resolved default is persisted to the
// session so it sticks across requests. Returns "" when the account has multiple
// orgs and none has been selected yet.
func (s *Server) resolveOrg(acct *app.Account, sessionID, headerOrg string) string {
	if sessionID != "" {
		s.mu.RLock()
		sess, ok := s.sessions[sessionID]
		s.mu.RUnlock()
		if ok && sess.OrgID != "" && accountHasOrgAccess(acct, sess.OrgID) {
			return sess.OrgID
		}
	}

	var orgID string
	switch {
	case headerOrg != "" && accountHasOrgAccess(acct, headerOrg):
		orgID = headerOrg
	case len(acct.OrgIDs) == 1:
		orgID = acct.OrgIDs[0]
	}

	if orgID != "" {
		s.setSessionOrg(sessionID, acct.ID, orgID)
	}
	return orgID
}

// setSessionOrg persists the selected org for an MCP session.
func (s *Server) setSessionOrg(sessionID, accountID, orgID string) {
	if sessionID == "" {
		return
	}
	s.mu.Lock()
	s.sessions[sessionID] = &authResult{OrgID: orgID, AccountID: accountID, lastSeen: time.Now()}
	s.mu.Unlock()
}

// touchSession refreshes a session's last-seen time so active sessions aren't
// evicted by the janitor.
func (s *Server) touchSession(sessionID string) {
	if sessionID == "" {
		return
	}
	s.mu.Lock()
	if sess, ok := s.sessions[sessionID]; ok {
		sess.lastSeen = time.Now()
	}
	s.mu.Unlock()
}

// runSessionJanitor periodically evicts idle sessions so the map stays bounded.
func (s *Server) runSessionJanitor() {
	ticker := time.NewTicker(mcpSessionTTL / 3)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopJanitor:
			return
		case <-ticker.C:
			s.evictStaleSessions()
		}
	}
}

func (s *Server) evictStaleSessions() {
	cutoff := time.Now().Add(-mcpSessionTTL)
	s.mu.Lock()
	for id, sess := range s.sessions {
		if sess.lastSeen.Before(cutoff) {
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()
}

func (s *Server) authContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Every request must carry a valid access token. A token failure returns
		// 401 + WWW-Authenticate so the client can discover the authorization
		// server and start the OAuth flow.
		acct, tokenRole, err := s.authenticateToken(r)
		if err != nil {
			s.l.Warn("MCP auth failed", zap.Error(err))
			s.writeUnauthorized(w, r)
			return
		}

		sessionID := r.Header.Get("Mcp-Session-Id")
		orgID := s.resolveOrg(acct, sessionID, r.Header.Get("X-Nuon-Org-ID"))
		s.touchSession(sessionID)

		ctx := withMCPAuth(r.Context(), orgID, acct.ID)
		ctx = keys.WithTokenRole(ctx, tokenRole)
		// Let the select_org tool change this session's active org.
		ctx = keys.WithOrgSelector(ctx, func(newOrgID string) {
			s.setSessionOrg(sessionID, acct.ID, newOrgID)
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withMCPAuth(ctx context.Context, orgID, accountID string) context.Context {
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, orgID)
	ctx = context.WithValue(ctx, keys.AccountIDCtxKey, accountID)
	return ctx
}
