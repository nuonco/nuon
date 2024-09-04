package middlewares

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	runnerCtxKey   string = "runner"
	runnerIDCtxKey string = "runner_id"
)

func RunnerIDFromContext(ctx context.Context) (string, error) {
	runner, err := RunnerFromContext(ctx)
	if err != nil {
		return "", err
	}

	return runner.ID, nil
}

func RunnerFromContext(ctx context.Context) (*app.Runner, error) {
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

func SetRunnerIDGinContext(ctx *gin.Context, runnerID string) {
	ctx.Set(runnerCtxKey, runnerID)
}

func SetRunnerContext(ctx context.Context, runner *app.Runner) context.Context {
	ctx = context.WithValue(ctx, runnerIDCtxKey, runner.ID)
	return context.WithValue(ctx, runnerCtxKey, runner)
}

func SetRunnerIDContext(ctx context.Context, runnerID string) context.Context {
	return context.WithValue(ctx, runnerIDCtxKey, runnerID)
}
