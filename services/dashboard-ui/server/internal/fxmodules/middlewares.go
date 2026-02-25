package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/ginmw"
	corsmw "github.com/nuonco/nuon/services/dashboard-ui/server/internal/middlewares/cors"
)

var MiddlewaresModule = fx.Module("middlewares",
	fx.Provide(ginmw.AsMiddleware(corsmw.New)),
)
