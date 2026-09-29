package signal

import (
	"errors"

	"go.temporal.io/sdk/temporal"
)

func HumanError(err error) string {
	if err == nil {
		return ""
	}

	var best string
	current := err
	for current != nil {
		var appErr *temporal.ApplicationError
		if errors.As(current, &appErr) {
			best = appErr.Message()
		}
		current = errors.Unwrap(current)
	}

	if best != "" {
		return best
	}

	return err.Error()
}
