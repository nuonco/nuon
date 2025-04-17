package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	runnerCtxKey string = "runner"
)

func RunnerFromContext(ctx ValueContext) (*app.Runner, error) {
	runner := ctx.Value(runnerCtxKey)
	if runner == nil {
		return nil, fmt.Errorf("runner was not set on middleware context")
	}

	return runner.(*app.Runner), nil
}

func SetRunnerGinContext(ctx *gin.Context, runner *app.Runner) {
	ctx.Set(runnerCtxKey, runner)
	ctx.Set(runnerIDCtxKey, runner.ID)
}

func SetRunnerContext(ctx context.Context, runner *app.Runner) context.Context {
	ctx = context.WithValue(ctx, runnerIDCtxKey, runner.ID)
	return context.WithValue(ctx, runnerCtxKey, runner)
}
