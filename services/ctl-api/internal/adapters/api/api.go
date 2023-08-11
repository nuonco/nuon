package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service interface {
	RegisterRoutes(*gin.Engine) error
	RegisterInternalRoutes(*gin.Engine) error
}

func AsService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Service)),
		fx.ResultTags(`group:"services"`),
	)
}

type API struct {
	public     *gin.Engine
	publicAddr string

	internal     *gin.Engine
	internalAddr string
}

func NewAPI(services []Service, lc fx.Lifecycle, l *zap.Logger, cfg *internal.Config) (*API, error) {
	api := &API{
		public:       gin.Default(),
		publicAddr:   fmt.Sprintf(":%v", cfg.HTTPPort),
		internal:     gin.Default(),
		internalAddr: fmt.Sprintf(":%v", cfg.InternalHTTPPort),
	}

	for idx, svc := range services {
		if err := svc.RegisterRoutes(api.public); err != nil {
			return nil, fmt.Errorf("unable to register routes on svc %d: %w", idx, err)
		}

		if err := svc.RegisterInternalRoutes(api.internal); err != nil {
			return nil, fmt.Errorf("unable to register internal routes on svc %d: %w", idx, err)
		}
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
