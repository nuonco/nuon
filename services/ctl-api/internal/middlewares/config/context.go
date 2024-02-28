package config

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

const (
	ContextKey string = "config"
)

var ErrConfigContextNotFound error = fmt.Errorf("config context not found")

func FromContext(ctx context.Context) (*internal.Config, error) {
	cfg := ctx.Value(ContextKey)
	if cfg == nil {
		return nil, ErrConfigContextNotFound

	}

	return cfg.(*internal.Config), nil
}
