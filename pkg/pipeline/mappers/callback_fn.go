package mappers

import (
	"context"
	"fmt"
)

// GetCallbackFn: returns a callback function
func (d *defaultMapper) GetCallbackFn(ctx context.Context, fn interface{}) (CallbackFn, error) {
	switch f := fn.(type) {
	case callbackNoop:
		return f.callback, nil
	default:
	}

	return nil, fmt.Errorf("unable to map to callback function: %T", fn)
}
