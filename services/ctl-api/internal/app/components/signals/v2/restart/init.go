package restart

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "component-restart"

func init() {
	catalog.Register(SignalType, func() signal.Signal {
		return &Signal{}
	})
}
