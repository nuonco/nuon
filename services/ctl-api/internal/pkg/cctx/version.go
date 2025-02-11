package cctx

import (
	"go.temporal.io/sdk/workflow"
)

const (
	versionCtxKey string = "version"
)

func SetVersionWorkflowContext(ctx workflow.Context, version string) workflow.Context {
	return workflow.WithValue(ctx, versionCtxKey, version)
}

func GetVersionWorkflowContext(ctx workflow.Context) string {
	val := ctx.Value(versionCtxKey)
	if val == nil {
		return ""
	}

	return val.(string)
}
