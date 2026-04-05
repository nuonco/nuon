package catalog

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

var SignalCatalog map[signal.SignalType]func() signal.Signal = make(map[signal.SignalType]func() signal.Signal, 0)

func Register(typ signal.SignalType, fn func() signal.Signal) {
	if _, exists := SignalCatalog[typ]; exists {
		panic(fmt.Sprintf("duplicate signal type registered: %q", typ))
	}
	SignalCatalog[typ] = fn
}
