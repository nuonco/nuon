package cctx

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func SetLogStreamContext(ctx context.Context, ls *app.LogStream) context.Context {
	return context.WithValue(ctx, keys.LogStreamCtxKey, ls)
}

func ClearLogStreamContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, keys.LogStreamCtxKey, struct{}{})
}

func GetLogStreamContext(ctx ValueContext) (*app.LogStream, error) {
	ls, ok := ctx.Value(keys.LogStreamCtxKey).(*app.LogStream)
	if !ok || ls == nil {
		return nil, fmt.Errorf("log stream not set on context")
	}

	return ls, nil
}

func SetLogStreamWorkflowContext(ctx workflow.Context, ls *app.LogStream) workflow.Context {
	return workflow.WithValue(ctx, keys.LogStreamCtxKey, ls)
}

func ClearLogStreamWorkflowContext(ctx workflow.Context) workflow.Context {
	return workflow.WithValue(ctx, keys.LogStreamCtxKey, struct{}{})
}

func GetLogStreamIDWorkflow(ctx ValueContext) (string, error) {
	ls, err := GetLogStreamWorkflow(ctx)
	if err != nil {
		return "", err
	}

	return ls.ID, nil
}

func GetLogStreamWorkflow(ctx ValueContext) (*app.LogStream, error) {
	ls, ok := ctx.Value(keys.LogStreamCtxKey).(*app.LogStream)
	if !ok || ls == nil {
		return nil, fmt.Errorf("no log stream found")
	}

	return ls, nil
}
