package fxmodules

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/ginmw"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/spa"
)

type APIParams struct {
	fx.In

	Config      *internal.Config
	Logger      *zap.Logger
	Middlewares []ginmw.Middleware `group:"middlewares"`
	Services    []Service          `group:"services"`
	SPA         *spa.Handler
}

type API struct {
	cfg         *internal.Config
	l           *zap.Logger
	middlewares []ginmw.Middleware
	services    []Service
	spa         *spa.Handler
	handler     *gin.Engine
	srv         *http.Server
}

func findAvailablePort(preferred string) (string, error) {
	port, _ := strconv.Atoi(preferred)
	base := 4000
	if port >= base {
		base = (port / 10) * 10
	}
	for p := base; p <= 4090; p += 10 {
		ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", p))
		if err == nil {
			ln.Close()
			return strconv.Itoa(p), nil
		}
	}
	return "", fmt.Errorf("no available ports in range 4000-4090")
}

func writePortFile(distDir, port string) error {
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(distDir, ".port"), []byte(port), 0o644)
}

func NewAPI(p APIParams) (*API, error) {
	port, err := findAvailablePort(p.Config.HTTPPort)
	if err != nil {
		return nil, fmt.Errorf("unable to find available port: %w", err)
	}
	if port != p.Config.HTTPPort {
		p.Logger.Info("configured port in use, selected alternative",
			zap.String("configured", p.Config.HTTPPort),
			zap.String("selected", port))
	}
	p.Config.HTTPPort = port
	p.Config.AppUrl = fmt.Sprintf("http://localhost:%s", port)

	if err := writePortFile(p.Config.DistDir, port); err != nil {
		p.Logger.Warn("failed to write port file", zap.Error(err))
	}

	handler := gin.New()
	handler.Use(gin.Recovery())
	handler.Use(gin.Logger())

	api := &API{
		cfg:         p.Config,
		l:           p.Logger,
		middlewares: p.Middlewares,
		services:    p.Services,
		spa:         p.SPA,
		handler:     handler,
		srv: &http.Server{
			Addr:    fmt.Sprintf("0.0.0.0:%s", port),
			Handler: handler.Handler(),
		},
	}

	if err := api.registerMiddlewares(); err != nil {
		return nil, fmt.Errorf("unable to register middlewares: %w", err)
	}

	if err := api.registerServices(); err != nil {
		return nil, fmt.Errorf("unable to register services: %w", err)
	}

	// SPA routes MUST be registered last — they use NoRoute as a catch-all
	// fallback for client-side routing.
	if err := api.spa.RegisterRoutes(api.handler); err != nil {
		return nil, fmt.Errorf("unable to register SPA routes: %w", err)
	}

	return api, nil
}

func (a *API) registerMiddlewares() error {
	lookup := make(map[string]gin.HandlerFunc, len(a.middlewares))
	for _, mw := range a.middlewares {
		lookup[mw.Name()] = mw.Handler()
	}

	for _, name := range a.cfg.Middlewares {
		fn, ok := lookup[name]
		if !ok {
			a.l.Warn("middleware not found, skipping", zap.String("name", name))
			continue
		}
		a.l.Info("registering middleware", zap.String("name", name))
		a.handler.Use(fn)
	}

	return nil
}

func (a *API) registerServices() error {
	for _, svc := range a.services {
		if err := svc.RegisterRoutes(a.handler); err != nil {
			return fmt.Errorf("unable to register routes: %w", err)
		}
	}
	return nil
}

func (a *API) lifecycleHooks(shutdowner fx.Shutdowner) fx.Hook {
	return fx.Hook{
		OnStart: func(_ context.Context) error {
			a.l.Info("starting dashboard BFF server", zap.String("addr", a.srv.Addr))
			go func() {
				if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					a.l.Error("server error", zap.Error(err))
					shutdowner.Shutdown(fx.ExitCode(127))
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			a.l.Info("stopping dashboard BFF server")
			return a.srv.Shutdown(context.Background())
		},
	}
}

var APIModule = fx.Module("api",
	fx.Provide(spa.NewHandler),
	fx.Provide(NewAPI),
	fx.Invoke(func(lc fx.Lifecycle, api *API, shutdowner fx.Shutdowner) {
		lc.Append(api.lifecycleHooks(shutdowner))
	}),
)
