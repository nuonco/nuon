package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
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
