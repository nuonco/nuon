package cctx

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

// Copy all known context fields from a workflow context, into a regular context
func ContextFromWorkflowContext(ctx context.Context, wCtx workflow.Context) context.Context {
	acctID, _ := AccountIDFromWorkflowContext(wCtx)
	orgID, _ := OrgIDFromWorkflowContext(wCtx)
	ls, _ := GetLogStreamWorkflow(wCtx)

	ctx = SetAccountIDContext(ctx, acctID)
	ctx = SetOrgIDContext(ctx, orgID)
	ctx = SetLogStreamContext(ctx, ls)

	return ctx
}

// Copy all known context fields from a workflow context, into a workflow context
func WorkflowContextFromContext(wCtx workflow.Context, ctx context.Context) workflow.Context {
	acctID, _ := AccountIDFromContext(ctx)
	orgID, _ := OrgIDFromContext(ctx)
	ls, _ := GetLogStreamContext(ctx)

	wCtx = SetAccountIDWorkflowContext(wCtx, acctID)
	wCtx = SetOrgIDWorkflowContext(wCtx, orgID)
	wCtx = SetLogStreamWorkflowContext(wCtx, ls)

	return nil
}
