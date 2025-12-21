package queues

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/queues/service"
)

var Module = fx.Module(
	"queues",
	fx.Provide(service.New),
)
