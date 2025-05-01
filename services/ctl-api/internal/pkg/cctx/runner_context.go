package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func RunnerFromContext(ctx ValueContext) (*app.Runner, error) {
	runner := ctx.Value(keys.RunnerCtxKey)
	if runner == nil {
		return nil, fmt.Errorf("runner was not set on middleware context")
	}

	return runner.(*app.Runner), nil
}

func SetRunnerGinContext(ctx *gin.Context, runner *app.Runner) {
	ctx.Set(keys.RunnerCtxKey, runner)
	ctx.Set(keys.RunnerIDCtxKey, runner.ID)
}

func SetRunnerContext(ctx context.Context, runner *app.Runner) context.Context {
	ctx = context.WithValue(ctx, keys.RunnerIDCtxKey, runner.ID)
	return context.WithValue(ctx, keys.RunnerCtxKey, runner)
}
