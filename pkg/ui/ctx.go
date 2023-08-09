package ui

import (
	"context"
	"fmt"
)

type uiContextKey struct{}

// ui exposes methods for outputting state during the command
func WithContext(ctx context.Context, log *logger) context.Context {
	return context.WithValue(ctx, uiContextKey{}, log)
}

// from context returns the logger from the ui
func FromContext(ctx context.Context) (*logger, error) {
	val := ctx.Value(uiContextKey{})
	if val == nil {
		return nil, fmt.Errorf("unable to get logger from context")
	}

	log, ok := val.(*logger)
	if !ok {
		return nil, fmt.Errorf("invalid object in context")
	}

	return log, nil
}
