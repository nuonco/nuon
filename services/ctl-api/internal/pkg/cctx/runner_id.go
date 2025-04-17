package cctx

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	runnerIDCtxKey string = "runner_id"
)

func RunnerIDFromContext(ctx ValueContext) (string, error) {
	runner, err := RunnerFromContext(ctx)
	if err != nil {
		return "", err
	}

	return runner.ID, nil
}

func SetRunnerIDGinContext(ctx *gin.Context, runnerID string) {
	ctx.Set(runnerCtxKey, runnerID)
}

func SetRunnerIDContext(ctx context.Context, runnerID string) context.Context {
	return context.WithValue(ctx, runnerIDCtxKey, runnerID)
}
