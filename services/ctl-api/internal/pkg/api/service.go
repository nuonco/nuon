package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

//

type Service interface {
	RegisterPublicRoutes(*gin.Engine) error
	RegisterRunnerRoutes(*gin.Engine) error
	RegisterInternalRoutes(*gin.Engine) error
	RegisterAuthRoutes(*gin.Engine) error
	RegisterAdminDashboardRoutes(*gin.Engine) error
	RegisterSlackRoutes(*gin.Engine) error
}

func AsService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Service)),
		fx.ResultTags(`group:"services"`),
	)
}
