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

// orgSelectionTTL bounds how long an idle org selection is retained.
const orgSelectionTTL = 30 * time.Minute

// orgSelection is the active org chosen for a token via select_org.
type orgSelection struct {
	orgID    string
	lastSeen time.Time
}

type Server struct {
	db          *gorm.DB
	l           *zap.Logger
	cfg         *internal.Config
	services    []api.Service
	httpServer  *http.Server
	schemaCache *mcp.SchemaCache

	mu            sync.RWMutex
	orgSelections map[string]*orgSelection
	stopJanitor   chan struct{}
}

func New(params Params) *Server {
	s := &Server{
		db:            params.DB,
		l:             params.L.Named("mcp"),
		cfg:           params.Cfg,
		services:      params.Services,
		schemaCache:   mcp.NewSchemaCache(),
		orgSelections: make(map[string]*orgSelection),
		stopJanitor:   make(chan struct{}),
	}

	mcpHandler := s.newMCPHandler()

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
			go s.runSelectionJanitor()
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

// newMCPHandler builds the streamable-HTTP handler. The server is stateless: it
// neither reads nor sets Mcp-Session-Id, and each request gets a temporary
// session. Everything a request needs is carried by its bearer token, so any
// replica can serve any request.
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
		Name:    "nuon-ctl",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		SchemaCache:  s.schemaCache,
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
// for this token wins; otherwise fall back to a valid X-Nuon-Org-ID header, or
// auto-select the account's only org. The resolved default is remembered so it
// sticks across requests. Returns "" when the account has multiple orgs and none
// has been selected yet.
//
// Selections are keyed by token rather than by Mcp-Session-Id because a
// stateless server has no session identifier to key on.
func (s *Server) resolveOrg(acct *app.Account, tokenID, headerOrg string) string {
	if tokenID != "" {
		s.mu.RLock()
		sel, ok := s.orgSelections[tokenID]
		s.mu.RUnlock()
		if ok && sel.orgID != "" && accountHasOrgAccess(acct, sel.orgID) {
			return sel.orgID
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
		s.setOrgSelection(tokenID, orgID)
	}
	return orgID
}

// setOrgSelection records the active org for a token.
func (s *Server) setOrgSelection(tokenID, orgID string) {
	if tokenID == "" {
		return
	}
	s.mu.Lock()
	s.orgSelections[tokenID] = &orgSelection{orgID: orgID, lastSeen: time.Now()}
	s.mu.Unlock()
}

// touchOrgSelection refreshes a selection's last-seen time so tokens in active
// use aren't evicted by the janitor.
func (s *Server) touchOrgSelection(tokenID string) {
	if tokenID == "" {
		return
	}
	s.mu.Lock()
	if sel, ok := s.orgSelections[tokenID]; ok {
		sel.lastSeen = time.Now()
	}
	s.mu.Unlock()
}

// runSelectionJanitor periodically evicts idle selections so the map stays
// bounded.
func (s *Server) runSelectionJanitor() {
	ticker := time.NewTicker(orgSelectionTTL / 3)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopJanitor:
			return
		case <-ticker.C:
			s.evictStaleOrgSelections()
		}
	}
}

func (s *Server) evictStaleOrgSelections() {
	cutoff := time.Now().Add(-orgSelectionTTL)
	s.mu.Lock()
	for id, sel := range s.orgSelections {
		if sel.lastSeen.Before(cutoff) {
			delete(s.orgSelections, id)
		}
	}
	s.mu.Unlock()
}

func (s *Server) authContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Every request must carry a valid access token. A token failure returns
		// 401 + WWW-Authenticate so the client can discover the authorization
		// server and start the OAuth flow.
		acct, tok, err := s.authenticateToken(r)
		if err != nil {
			s.l.Warn("MCP auth failed", zap.Error(err))
			s.writeUnauthorized(w, r)
			return
		}

		orgID := s.resolveOrg(acct, tok.ID, r.Header.Get("X-Nuon-Org-ID"))
		s.touchOrgSelection(tok.ID)

		ctx := withMCPAuth(r.Context(), orgID, acct.ID)
		ctx = keys.WithTokenRole(ctx, tok.Role)
		// Let the select_org tool change the active org for this token.
		ctx = keys.WithOrgSelector(ctx, func(newOrgID string) {
			s.setOrgSelection(tok.ID, newOrgID)
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withMCPAuth(ctx context.Context, orgID, accountID string) context.Context {
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, orgID)
	ctx = context.WithValue(ctx, keys.AccountIDCtxKey, accountID)
	return ctx
}
