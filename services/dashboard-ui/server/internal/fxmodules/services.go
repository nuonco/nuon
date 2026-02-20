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
	fx.Provide(asService(handlers.NewAccountHandler)),
	fx.Provide(asService(handlers.NewOrgsHandler)),
	fx.Provide(asService(handlers.NewAppsHandler)),
	fx.Provide(asService(handlers.NewInstallsHandler)),
	fx.Provide(asService(handlers.NewComponentsHandler)),
	fx.Provide(asService(handlers.NewRunnersHandler)),
	fx.Provide(asService(handlers.NewWorkflowsHandler)),
	fx.Provide(asService(handlers.NewVCSHandler)),
	fx.Provide(asService(handlers.NewLogStreamsHandler)),
	fx.Provide(asService(handlers.NewActionsHandler)),
	fx.Provide(asService(handlers.NewSSEHandler)),
	fx.Provide(asService(handlers.NewProxyHandler)),
)
