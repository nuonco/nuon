package cctx

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)


func RunnerIDFromContext(ctx ValueContext) (string, error) {
	runner, err := RunnerFromContext(ctx)
	if err != nil {
		return "", err
	}

	return runner.ID, nil
}

func SetRunnerIDGinContext(ctx *gin.Context, runnerID string) {
	ctx.Set(keys.RunnerIDCtxKey, runnerID)
}

func SetRunnerIDContext(ctx context.Context, runnerID string) context.Context {
	return context.WithValue(ctx, keys.RunnerIDCtxKey, runnerID)
}
