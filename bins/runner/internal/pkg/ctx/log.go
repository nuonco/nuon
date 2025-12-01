package pkgctx

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

var errLoggerNotFound error = fmt.Errorf("logger not found in context")

type logCtxKey struct{}

func Logger(ctx context.Context) (*zap.Logger, error) {
	val := ctx.Value(logCtxKey{})
	if val == nil {
		return nil, errLoggerNotFound
	}

	return val.(*zap.Logger), nil
}

func SetLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, logCtxKey{}, l)
}
