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

type Server struct {
	db         *gorm.DB
	l          *zap.Logger
	cfg        *internal.Config
	services   []api.Service
	httpServer *http.Server

	mu       sync.RWMutex
	sessions map[string]*authResult
}

func New(params Params) *Server {
	s := &Server{
		db:       params.DB,
		l:        params.L.Named("mcp"),
		cfg:      params.Cfg,
		services: params.Services,
		sessions: make(map[string]*authResult),
	}

	mcpHandler := mcp.NewStreamableHTTPHandler(
		s.getServerForRequest,
		&mcp.StreamableHTTPOptions{
			SessionTimeout: 30 * time.Minute,
			Logger:         slog.Default(),
		},
	)

	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", params.Cfg.MCPHTTPPort),
		Handler: s.authContextMiddleware(mcpHandler),
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
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.l.Info("stopping MCP server")
			return s.httpServer.Shutdown(ctx)
		},
	})

	return s
}

func (s *Server) getServerForRequest(r *http.Request) *mcp.Server {
	auth, err := s.authenticate(r)
	if err != nil {
		s.l.Warn("MCP auth failed", zap.Error(err))
		return nil
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nuon-ctl",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Instructions: fmt.Sprintf("Nuon control plane MCP server. Authenticated as account %s in org %s.", auth.AccountID, auth.OrgID),
	})

	for _, svc := range s.services {
		if mcpSvc, ok := svc.(api.MCPService); ok {
			mcpSvc.RegisterMCPTools(server)
		}
	}

	sessionID := r.Header.Get("Mcp-Session-Id")
	if sessionID != "" {
		s.mu.Lock()
		s.sessions[sessionID] = auth
		s.mu.Unlock()
	}

	return server
}

func (s *Server) authContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("Mcp-Session-Id")
		if sessionID != "" {
			s.mu.RLock()
			auth, ok := s.sessions[sessionID]
			s.mu.RUnlock()
			if ok {
				ctx := withMCPAuth(r.Context(), auth.OrgID, auth.AccountID)
				r = r.WithContext(ctx)
			}
		}

		if keys.OrgIDFromContext(r.Context()) == "" {
			auth, err := s.authenticate(r)
			if err == nil {
				ctx := withMCPAuth(r.Context(), auth.OrgID, auth.AccountID)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func withMCPAuth(ctx context.Context, orgID, accountID string) context.Context {
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, orgID)
	ctx = context.WithValue(ctx, keys.AccountIDCtxKey, accountID)
	return ctx
}
