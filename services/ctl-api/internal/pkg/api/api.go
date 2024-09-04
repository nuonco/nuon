package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type API struct {
	public     *gin.Engine
	publicAddr string

	internal     *gin.Engine
	internalAddr string

	runner     *gin.Engine
	runnerAddr string

	services    []Service
	middlewares []Middleware
	l           *zap.Logger
	cfg         *internal.Config
}

func NewAPI(services []Service,
	middlewares []Middleware,
	lc fx.Lifecycle,
	l *zap.Logger,
	cfg *internal.Config,
	shutdowner fx.Shutdowner,
) (*API, error) {
	api := &API{
		public:       gin.Default(),
		publicAddr:   fmt.Sprintf(":%v", cfg.HTTPPort),
		internal:     gin.Default(),
		internalAddr: fmt.Sprintf(":%v", cfg.InternalHTTPPort),
		runner:       gin.Default(),
		runnerAddr:   fmt.Sprintf(":%v", cfg.RunnerHTTPPort),

		cfg:         cfg,
		services:    services,
		middlewares: middlewares,
		l:           l,
	}

	if err := api.registerMiddlewares(); err != nil {
		return nil, fmt.Errorf("unable to register middlewares: %w", err)
	}
	if err := api.registerServices(); err != nil {
		return nil, fmt.Errorf("unable to register middlewares: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			l.Info("starting public api", zap.String("addr", api.publicAddr))
			go func() {
				if err := api.public.Run(api.publicAddr); err != nil {
					l.Error("unable to run public api", zap.Error(err))
					shutdowner.Shutdown(fx.ExitCode(127))
				}
			}()

			l.Info("starting internal api", zap.String("addr", api.internalAddr))
			go func() {
				if err := api.internal.Run(api.internalAddr); err != nil {
					l.Error("unable to run internal api", zap.Error(err))
					shutdowner.Shutdown(fx.ExitCode(127))
				}
			}()

			l.Info("starting runner api", zap.String("addr", api.runnerAddr))
			go func() {
				if err := api.runner.Run(api.runnerAddr); err != nil {
					l.Error("unable to run runner api", zap.Error(err))
					shutdowner.Shutdown(fx.ExitCode(127))
				}
			}()

			return nil
		},
		OnStop: func(_ context.Context) error {
			l.Info("stopping public api")
			return nil
		},
	})

	return api, nil
}
