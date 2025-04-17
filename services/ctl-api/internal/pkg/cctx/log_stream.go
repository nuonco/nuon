package cctx

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	logStreamCtxKey string = "log_stream"
)

func SetLogStreamContext(ctx context.Context, ls *app.LogStream) context.Context {
	return context.WithValue(ctx, logStreamCtxKey, ls)
}

func GetLogStreamContext(ctx ValueContext) (*app.LogStream, error) {
	ls := ctx.Value(logStreamCtxKey)
	if ls == nil {
		return nil, fmt.Errorf("log stream not set on context")
	}

	return ls.(*app.LogStream), nil
}

func SetLogStreamWorkflowContext(ctx workflow.Context, ls *app.LogStream) workflow.Context {
	return workflow.WithValue(ctx, logStreamCtxKey, ls)
}

func GetLogStreamIDWorkflow(ctx ValueContext) (string, error) {
	ls, err := GetLogStreamWorkflow(ctx)
	if err != nil {
		return "", err
	}

	return ls.ID, nil
}

func GetLogStreamWorkflow(ctx ValueContext) (*app.LogStream, error) {
	val := ctx.Value(logStreamCtxKey)
	if val == nil {
		return nil, fmt.Errorf("no log stream found")
	}

	return val.(*app.LogStream), nil
}
