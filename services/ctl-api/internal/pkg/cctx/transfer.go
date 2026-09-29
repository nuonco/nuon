package cctx

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

func ContextFromWorkflowContext(ctx context.Context, wCtx workflow.Context) context.Context {
	acctID, _ := AccountIDFromContext(wCtx)
	isEmployee, _ := IsEmployeeFromContext(wCtx)
	orgID, _ := OrgIDFromContext(wCtx)
	ls, _ := GetLogStreamWorkflow(wCtx)

	ctx = SetAccountIDContext(ctx, acctID)
	ctx = SetOrgIDContext(ctx, orgID)
	ctx = SetLogStreamContext(ctx, ls)
	ctx = SetIsEmployeeContext(ctx, isEmployee)

	return ctx
}

func WorkflowContextFromContext(wCtx workflow.Context, ctx ValueContext) workflow.Context {
	acctID, _ := AccountIDFromContext(ctx)
	orgID, _ := OrgIDFromContext(ctx)
	ls, _ := GetLogStreamContext(ctx)
	isEmployee, _ := IsEmployeeFromContext(ctx)

	wCtx = SetAccountIDWorkflowContext(wCtx, acctID)
	wCtx = SetOrgIDWorkflowContext(wCtx, orgID)
	wCtx = SetLogStreamWorkflowContext(wCtx, ls)
	wCtx = SetIsEmployeeWorkflowContext(wCtx, isEmployee)

	return wCtx
}
