package cctx

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func SetInstallWorkflowContext(ctx context.Context, wf *app.InstallWorkflowContext) context.Context {
	return context.WithValue(ctx, keys.WorkflowCtxKey, wf)
}

func GetInstallWorkflowContext(ctx ValueContext) (*app.InstallWorkflowContext, error) {
	wf := ctx.Value(keys.WorkflowCtxKey)
	if wf == nil {
		return nil, fmt.Errorf("workflow not set on context")
	}

	return wf.(*app.InstallWorkflowContext), nil
}

func SetInstallWorkflowWorkflowContext(ctx workflow.Context, wf *app.InstallWorkflowContext) workflow.Context {
	return workflow.WithValue(ctx, keys.WorkflowCtxKey, wf)
}

func GetInstallWorkflowIDWorkflow(ctx ValueContext) (string, error) {
	wf, err := GetInstallWorkflowWorkflow(ctx)
	if err != nil {
		return "", err
	}

	return wf.ID, nil
}

func GetInstallWorkflowWorkflow(ctx ValueContext) (*app.InstallWorkflowContext, error) {
	val := ctx.Value(keys.WorkflowCtxKey)
	if val == nil {
		return nil, fmt.Errorf("workflow context not found")
	}

	return val.(*app.InstallWorkflowContext), nil
}
