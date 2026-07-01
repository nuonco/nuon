package cctx

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func FlowWorkflowIDFromContext(ctx ValueContext) string {
	return stringValue(ctx, keys.FlowWorkflowIDCtxKey)
}
func FlowStepIDFromContext(ctx ValueContext) string { return stringValue(ctx, keys.FlowStepIDCtxKey) }
func FlowInstallIDFromContext(ctx ValueContext) string {
	return stringValue(ctx, keys.FlowInstallIDCtxKey)
}

func stringValue(ctx ValueContext, key string) string {
	v := ctx.Value(key)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func SetFlowWorkflowIDContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keys.FlowWorkflowIDCtxKey, id)
}
func SetFlowStepIDContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keys.FlowStepIDCtxKey, id)
}
func SetFlowInstallIDContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keys.FlowInstallIDCtxKey, id)
}

func SetFlowWorkflowIDWorkflowContext(ctx workflow.Context, id string) workflow.Context {
	return workflow.WithValue(ctx, keys.FlowWorkflowIDCtxKey, id)
}
func SetFlowStepIDWorkflowContext(ctx workflow.Context, id string) workflow.Context {
	return workflow.WithValue(ctx, keys.FlowStepIDCtxKey, id)
}
func SetFlowInstallIDWorkflowContext(ctx workflow.Context, id string) workflow.Context {
	return workflow.WithValue(ctx, keys.FlowInstallIDCtxKey, id)
}
