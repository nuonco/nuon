package callbacks

import "context"

// Noop is a callback that can be used to noop
func Noop(_ context.Context) error {
	return nil
}
