package vcspush

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func init() {
	catalog.Register(SignalType, func() signal.Signal {
		return &Signal{}
	})
}
