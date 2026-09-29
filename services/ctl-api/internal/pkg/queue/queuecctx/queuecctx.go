package queuecctx

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	qcctx "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/cctx"
)

func FromContext(ctx cctx.ValueContext) qcctx.SignalContext {
	sc := qcctx.SignalContext{}

	sc.AccountID, _ = cctx.AccountIDFromContext(ctx)
	sc.OrgID, _ = cctx.OrgIDFromContext(ctx)
	sc.TraceID = cctx.TraceIDFromContext(ctx)
	sc.WorkflowTelemetry = cctx.WorkflowTelemetryFromContext(ctx)

	if ls, err := cctx.GetLogStreamContext(ctx); err == nil && ls != nil {
		sc.LogStreamID = ls.ID
	}

	return sc
}

func Apply(ctx context.Context, sc qcctx.SignalContext) context.Context {
	ctx = cctx.ClearLogStreamContext(ctx)
	if sc.AccountID != "" {
		ctx = cctx.SetAccountIDContext(ctx, sc.AccountID)
	}
	if sc.OrgID != "" {
		ctx = cctx.SetOrgIDContext(ctx, sc.OrgID)
	}
	if sc.TraceID != "" {
		ctx = cctx.SetTraceIDContext(ctx, sc.TraceID)
	}
	ctx = cctx.SetWorkflowTelemetryContext(ctx, sc.WorkflowTelemetry)
	return ctx
}

func ApplyWorkflow(ctx workflow.Context, sc qcctx.SignalContext) workflow.Context {
	ctx = cctx.ClearLogStreamWorkflowContext(ctx)
	if sc.AccountID != "" {
		ctx = cctx.SetAccountIDWorkflowContext(ctx, sc.AccountID)
	}
	if sc.OrgID != "" {
		ctx = cctx.SetOrgIDWorkflowContext(ctx, sc.OrgID)
	}
	if sc.TraceID != "" {
		ctx = cctx.SetTraceIDWorkflowContext(ctx, sc.TraceID)
	}
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, sc.WorkflowTelemetry)
	return ctx
}
