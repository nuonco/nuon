package cctx

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	workflowCtxKey string = "install_workflow"
)

func SetInstallWorkflowContext(ctx context.Context, wf *app.InstallWorkflowContext) context.Context {
	return context.WithValue(ctx, workflowCtxKey, wf)
}

func GetInstallWorkflowContext(ctx ValueContext) (*app.InstallWorkflowContext, error) {
	wf := ctx.Value(logStreamCtxKey)
	if wf == nil {
		return nil, fmt.Errorf("log stream not set on context")
	}

	return wf.(*app.InstallWorkflowContext), nil
}

func SetInstallWorkflowWorkflowContext(ctx workflow.Context, wf *app.InstallWorkflowContext) workflow.Context {
	return workflow.WithValue(ctx, workflowCtxKey, wf)
}

func GetInstallWorkflowIDWorkflow(ctx ValueContext) (string, error) {
	wf, err := GetInstallWorkflowWorkflow(ctx)
	if err != nil {
		return "", err
	}

	return wf.ID, nil
}

func GetInstallWorkflowWorkflow(ctx ValueContext) (*app.InstallWorkflowContext, error) {
	val := ctx.Value(workflowCtxKey)
	if val == nil {
		return nil, fmt.Errorf("no log stream found")
	}

	return val.(*app.InstallWorkflowContext), nil
}
