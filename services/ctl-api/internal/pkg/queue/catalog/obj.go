package catalog

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// ErrSignalTypeNotRegistered is returned by NewFromType when the given signal
// type is not present in the catalog. Callers can use errors.Is to detect this
// specific condition — e.g. to gracefully skip stale DB rows that reference
// deprecated signal types instead of failing the whole query.
var ErrSignalTypeNotRegistered = errors.New("signal type not registered")

func NewFromType(typ signal.SignalType) (signal.Signal, error) {
	constructor, ok := SignalCatalog[typ]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSignalTypeNotRegistered, typ)
	}

	return constructor(), nil
}
