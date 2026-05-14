package fxmodules

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/handlers"
)

// Service is the interface that handler groups implement to register routes.
type Service interface {
	RegisterRoutes(*gin.Engine) error
}

func asService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Service)),
		fx.ResultTags(`group:"services"`),
	)
}

var ServicesModule = fx.Module("services",
	fx.Provide(asService(handlers.NewHealthHandler)),
	fx.Provide(asService(handlers.NewRootHandler)),
	fx.Provide(asService(handlers.NewConnectHandler)),
	fx.Provide(asService(handlers.NewWorkflowsHandler)),
	fx.Provide(asService(handlers.NewLogStreamsHandler)),
	fx.Provide(asService(handlers.NewProxyHandler)),
	fx.Provide(asService(handlers.NewAPIProxyHandler)),
	fx.Provide(asService(handlers.NewRandomNamesHandler)),
	fx.Provide(asService(handlers.NewDDProxyHandler)),
	fx.Provide(asService(handlers.NewDeploysHandler)),
	fx.Provide(asService(handlers.NewSandboxRunsHandler)),
	fx.Provide(asService(handlers.NewActionRunsHandler)),
	fx.Provide(asService(handlers.NewBuildsHandler)),
	fx.Provide(asService(handlers.NewSandboxBuildsHandler)),
)
