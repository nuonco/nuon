package ginmw

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Middleware is the shared interface for Gin middlewares registered via FX.
// Both ctl-api and dashboard-ui BFF use this interface.
type Middleware interface {
	Name() string
	Handler() gin.HandlerFunc
}

// AsMiddleware annotates a constructor so FX collects it into the "middlewares" group.
func AsMiddleware(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Middleware)),
		fx.ResultTags(`group:"middlewares"`),
	)
}
