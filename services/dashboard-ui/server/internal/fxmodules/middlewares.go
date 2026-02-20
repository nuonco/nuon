package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/ginmw"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/middlewares/apiclient"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/middlewares/auth"
)

var MiddlewaresModule = fx.Module("middlewares",
	fx.Provide(ginmw.AsMiddleware(auth.New)),
	fx.Provide(ginmw.AsMiddleware(apiclient.New)),
)
