package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
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

func AsAPI(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"api"`),
	)
}

func APIGroupParam(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"api"`),
	)
}
