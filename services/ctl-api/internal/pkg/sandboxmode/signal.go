package sandboxmode

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Signal struct {
	signal.Signal
}

func WrapSignal(inner signal.Signal) *Signal {
	return &Signal{Signal: inner}
}
