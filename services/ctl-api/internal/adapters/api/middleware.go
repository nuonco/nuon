package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Middleware interface {
	Name() string
	Handler() gin.HandlerFunc
}

func AsMiddleware(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Middleware)),
		fx.ResultTags(`group:"middlewares"`),
	)
}

func (a *API) registerMiddlewares() error {
	// register middlewares
	middlewaresLookup := make(map[string]gin.HandlerFunc, 0)
	for _, middleware := range a.middlewares {
		middlewaresLookup[middleware.Name()] = middleware.Handler()
	}

	for _, middleware := range a.cfg.Middlewares {
		a.l.Info("middleware", zap.String("name", middleware))
		fn, ok := middlewaresLookup[middleware]
		if !ok {
			return fmt.Errorf("middleware not found: %s", middleware)
		}
		a.public.Use(fn)
	}
	for _, middleware := range a.cfg.InternalMiddlewares {
		fn, ok := middlewaresLookup[middleware]
		if !ok {
			return fmt.Errorf("middleware not found: %s", middleware)
		}
		a.internal.Use(fn)
	}

	return nil
}
