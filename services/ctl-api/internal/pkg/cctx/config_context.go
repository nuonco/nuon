package cctx

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

const (
	cfgCtxKey string = "config"
)

var ErrConfigContextNotFound error = fmt.Errorf("config context not found")

func ConfigFromContext(ctx ValueContext) (*internal.Config, error) {
	cfg := ctx.Value(cfgCtxKey)
	if cfg == nil {
		return nil, ErrConfigContextNotFound
	}

	return cfg.(*internal.Config), nil
}

func SetConfigGinContext(ctx *gin.Context, cfg *internal.Config) {
	ctx.Set(cfgCtxKey, cfg)
}
