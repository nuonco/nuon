package catalog

import (
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

var ErrSignalTypeNotRegistered = errors.New("signal type not registered")

const SignalTypeNotRegisteredErrorType = "SignalTypeNotRegistered"

func NewFromType(typ signal.SignalType) (signal.Signal, error) {
	constructor, ok := SignalCatalog[typ]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSignalTypeNotRegistered, typ)
	}

	return constructor(), nil
}

func IsSignalTypeNotRegistered(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrSignalTypeNotRegistered) {
		return true
	}
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) && appErr.Type() == SignalTypeNotRegisteredErrorType {
		return true
	}
	return false
}
