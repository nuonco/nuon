package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
)

type API struct {
	services              []Service
	middlewares           []middlewares.Middleware
	l                     *zap.Logger
	cfg                   *internal.Config
	port                  string
	name                  string
	configuredMiddlewares []string
	endpointAudit         *EndpointAudit
	// created after initializing
	srv     *http.Server
	handler *gin.Engine

	db *gorm.DB
}

func (a *API) init(httpMetrics *metrics.HTTPMetrics) error {
	a.handler = gin.New()
	if httpMetrics != nil {
		a.handler.Use(func(c *gin.Context) {
			metrics.SetHTTPRoute(c.Request.Context(), c.FullPath())
		})
	}
	a.srv = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", a.port),
		Handler: a.handler.Handler(),
	}
	if httpMetrics != nil {
		a.srv.Handler = httpMetrics.Handler(a.name, a.srv.Handler)
	}

	return nil
}

func (a *API) middlewareDebugWrapper(name string, fn gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		a.l.Debug("starting middleware " + name)
		fn(ctx)
		a.l.Debug("finished middleware " + name)
	}
}

const panickerMiddlewareName = "panicker"

func (a *API) registerMiddlewares() error {
	// register middlewares
	middlewaresLookup := make(map[string]gin.HandlerFunc, 0)
	for _, middleware := range a.middlewares {
		middlewaresLookup[middleware.Name()] = middleware.Handler()
	}

	// gin.CustomRecovery only recovers what runs after it, and gin.New() installs no
	// recovery of its own, so a panic in an earlier middleware kills the connection
	// and the caller sees a proxy 503 with no body instead of an error response.
	ordered := make([]string, 0, len(a.configuredMiddlewares))
	if slices.Contains(a.configuredMiddlewares, panickerMiddlewareName) {
		ordered = append(ordered, panickerMiddlewareName)
	}
	for _, middleware := range a.configuredMiddlewares {
		if middleware != panickerMiddlewareName {
			ordered = append(ordered, middleware)
		}
	}

	for _, middleware := range ordered {
		a.l.Info(fmt.Sprintf("registering middleware: %s", middleware), zap.String("name", middleware))
		fn, ok := middlewaresLookup[middleware]
		if !ok {
			return fmt.Errorf("middleware not found: %s", middleware)
		}
		a.handler.Use(a.middlewareDebugWrapper(middleware, fn))
	}

	return nil
}

func (a *API) registerServices() error {
	// register services
	for _, svc := range a.services {
		method, ok := map[string]func(*gin.Engine) error{
			"runner":          svc.RegisterRunnerRoutes,
			"public":          svc.RegisterPublicRoutes,
			"internal":        svc.RegisterInternalRoutes,
			"auth":            svc.RegisterAuthRoutes,
			"admin-dashboard": svc.RegisterAdminDashboardRoutes,
			"slack":           svc.RegisterSlackRoutes,
		}[a.name]
		if !ok {
			return fmt.Errorf("%s", "invalid name "+a.name)
		}

		if err := method(a.handler); err != nil {
			return fmt.Errorf("unable to register routes: %w", err)
		}
	}

	return nil
}

func (a *API) lifecycleHooks(shutdowner fx.Shutdowner) fx.Hook {
	return fx.Hook{
		OnStart: func(_ context.Context) error {
			if err := a.start(shutdowner); err != nil {
				a.l.Error("error starting server", zap.Error(err), zap.String("name", a.name))
			}

			if a.cfg.EnableEndpointAuditing {

				routes := []app.EndpointAudit{}

				for _, route := range a.handler.Routes() {
					deprecated := a.endpointAudit.IsDeprecated(route.Method, a.name, route.Path)

					routes = append(routes, app.EndpointAudit{
						Method: route.Method,
						Name:   a.name,
						Route:  route.Path,
						// Deprecated: set to true if the route is deprecated
						Deprecated: deprecated,
					})
				}

				if len(routes) == 0 {
					a.l.Warn("endpoint auditing enabled but no routes found", zap.String("name", a.name))
					return nil
				}

				ctx := context.Background()
				if res := a.db.WithContext(ctx).
					Clauses(clause.OnConflict{
						Columns: []clause.Column{
							{Name: "deleted_at"},
							{Name: "method"},
							{Name: "name"},
							{Name: "route"},
						},
						DoUpdates: clause.AssignmentColumns([]string{
							"deprecated",
						}),
					}).
					Create(&routes); res.Error != nil {
					return errors.Wrap(res.Error, "unable to write routes")
				}
			}

			return nil
		},
		OnStop: func(_ context.Context) error {
			if err := a.shutdown(); err != nil {
				a.l.Error("error shutting down server", zap.Error(err), zap.String("name", a.name))
			}
			return nil
		},
	}
}

func (a *API) start(shutdowner fx.Shutdowner) error {
	a.l.Info(fmt.Sprintf("starting %s api", a.name), zap.String("addr", a.srv.Addr))

	go func() {
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.l.Error(fmt.Sprintf("unable to run %s api", a.name), zap.Error(err))
			shutdowner.Shutdown(fx.ExitCode(127))
		}
	}()

	return nil
}

func (a *API) shutdown() error {
	a.l.Info(fmt.Sprintf("gracefully stopping %s api", a.name), zap.String("addr", a.srv.Addr))
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GracefulShutdownTimeout)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "unable to shutdown "+a.name)
	}

	a.l.Info("successfully stopped all handlers")
	return nil
}
