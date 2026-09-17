package cctx

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func WorkflowTypeFromContext(ctx ValueContext) string {
	s, _ := ctx.Value(keys.FlowWorkflowTypeCtxKey).(string)
	return s
}

func OrgNameFromContext(ctx ValueContext) string {
	s, _ := ctx.Value(keys.FlowOrgNameCtxKey).(string)
	return s
}

func InstallNameFromContext(ctx ValueContext) string {
	s, _ := ctx.Value(keys.FlowInstallNameCtxKey).(string)
	return s
}

func SetWorkflowTypeContext(ctx context.Context, workflowType string) context.Context {
	if workflowType == "" {
		return ctx
	}
	return context.WithValue(ctx, keys.FlowWorkflowTypeCtxKey, workflowType)
}

func SetOrgNameContext(ctx context.Context, orgName string) context.Context {
	if orgName == "" {
		return ctx
	}
	return context.WithValue(ctx, keys.FlowOrgNameCtxKey, orgName)
}

func SetInstallNameContext(ctx context.Context, installName string) context.Context {
	if installName == "" {
		return ctx
	}
	return context.WithValue(ctx, keys.FlowInstallNameCtxKey, installName)
}

func SetWorkflowTypeWorkflowContext(ctx workflow.Context, workflowType string) workflow.Context {
	if workflowType == "" {
		return ctx
	}
	return workflow.WithValue(ctx, keys.FlowWorkflowTypeCtxKey, workflowType)
}

func SetOrgNameWorkflowContext(ctx workflow.Context, orgName string) workflow.Context {
	if orgName == "" {
		return ctx
	}
	return workflow.WithValue(ctx, keys.FlowOrgNameCtxKey, orgName)
}

func SetInstallNameWorkflowContext(ctx workflow.Context, installName string) workflow.Context {
	if installName == "" {
		return ctx
	}
	return workflow.WithValue(ctx, keys.FlowInstallNameCtxKey, installName)
}
