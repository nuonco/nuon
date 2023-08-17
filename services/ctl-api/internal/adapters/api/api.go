package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type API struct {
	public     *gin.Engine
	publicAddr string

	internal     *gin.Engine
	internalAddr string

	services    []Service
	middlewares []Middleware
	l           *zap.Logger
	cfg         *internal.Config
}

func NewAPI(services []Service, middlewares []Middleware, lc fx.Lifecycle, l *zap.Logger, cfg *internal.Config) (*API, error) {
	api := &API{
		public:       gin.Default(),
		publicAddr:   fmt.Sprintf(":%v", cfg.HTTPPort),
		internal:     gin.Default(),
		internalAddr: fmt.Sprintf(":%v", cfg.InternalHTTPPort),

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
			go api.public.Run(api.publicAddr)

			l.Info("starting internal api", zap.String("addr", api.internalAddr))
			go api.internal.Run(api.internalAddr)

			return nil
		},
		OnStop: func(_ context.Context) error {
			l.Info("stopping public api")
			return nil
		},
	})

	return api, nil
}
