package catalog

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

func NewFromType(typ signal.SignalType) (signal.Signal, error) {
	constructor, ok := SignalCatalog[typ]
	if !ok {
		return nil, fmt.Errorf("invalid signal type %s - not registered", typ)
	}

	return constructor(), nil
}
