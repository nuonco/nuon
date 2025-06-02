package cctx

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func SetInstallWorkflowContext(ctx context.Context, wf *app.InstallWorkflowContext) context.Context {
	return context.WithValue(ctx, keys.InstallWorkflowCtxKey, wf)
}

func GetInstallWorkflowContext(ctx ValueContext) (*app.InstallWorkflowContext, error) {
	wf := ctx.Value(keys.InstallWorkflowCtxKey)
	if wf == nil {
		return nil, fmt.Errorf("workflow not set on context")
	}

	return wf.(*app.InstallWorkflowContext), nil
}

func SetInstallWorkflowWorkflowContext(ctx workflow.Context, wf *app.InstallWorkflowContext) workflow.Context {
	return workflow.WithValue(ctx, keys.InstallWorkflowCtxKey, wf)
}

func GetInstallWorkflowIDWorkflow(ctx ValueContext) (string, error) {
	wf, err := GetInstallWorkflowWorkflow(ctx)
	if err != nil {
		return "", err
	}

	return wf.ID, nil
}

func GetInstallWorkflowWorkflow(ctx ValueContext) (*app.InstallWorkflowContext, error) {
	val := ctx.Value(keys.InstallWorkflowCtxKey)
	if val == nil {
		return nil, fmt.Errorf("workflow context not found")
	}

	return val.(*app.InstallWorkflowContext), nil
}

// TODO(sdboyer) remove everything above here after refactor

func SetFlowContext(ctx context.Context, wf *app.FlowContext) context.Context {
	return context.WithValue(ctx, keys.FlowCtxKey, wf)
}

func SetFlowContextWithinWorkflow(ctx workflow.Context, wf *app.FlowContext) workflow.Context {
	return workflow.WithValue(ctx, keys.FlowCtxKey, wf)
}

func GetFlowContext(ctx ValueContext) (*app.FlowContext, error) {
	wf := ctx.Value(keys.FlowCtxKey)
	if wf == nil {
		return nil, fmt.Errorf("workflow not set on context")
	}

	return wf.(*app.FlowContext), nil
}

func GetFlowID(ctx ValueContext) (string, error) {
	wf, err := GetFlowContext(ctx)
	if err != nil {
		return "", err
	}

	return wf.ID, nil
}
