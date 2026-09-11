package queuecctx

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	qcctx "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/cctx"
)

// FromContext reads the common context values and returns a SignalContext
// suitable for persisting alongside a QueueSignal record.
func FromContext(ctx cctx.ValueContext) qcctx.SignalContext {
	sc := qcctx.SignalContext{}

	sc.AccountID, _ = cctx.AccountIDFromContext(ctx)
	sc.OrgID, _ = cctx.OrgIDFromContext(ctx)
	sc.TraceID = cctx.TraceIDFromContext(ctx)

	if ls, err := cctx.GetLogStreamContext(ctx); err == nil && ls != nil {
		sc.LogStreamID = ls.ID
	}

	return sc
}

// Apply restores the captured context values onto a context.Context so that
// downstream consumers (e.g. Temporal propagators) see the original values.
func Apply(ctx context.Context, sc qcctx.SignalContext) context.Context {
	if sc.AccountID != "" {
		ctx = cctx.SetAccountIDContext(ctx, sc.AccountID)
	}
	if sc.OrgID != "" {
		ctx = cctx.SetOrgIDContext(ctx, sc.OrgID)
	}
	if sc.TraceID != "" {
		ctx = cctx.SetTraceIDContext(ctx, sc.TraceID)
	}
	return ctx
}

// ApplyWorkflow is the workflow.Context analog of Apply. It restores the
// captured per-signal context values onto a workflow.Context so that the
// signal's Execute (and any activities it schedules via the Temporal
// propagator) see the enqueuer's identity rather than the queue workflow's.
func ApplyWorkflow(ctx workflow.Context, sc qcctx.SignalContext) workflow.Context {
	if sc.AccountID != "" {
		ctx = cctx.SetAccountIDWorkflowContext(ctx, sc.AccountID)
	}
	if sc.OrgID != "" {
		ctx = cctx.SetOrgIDWorkflowContext(ctx, sc.OrgID)
	}
	if sc.TraceID != "" {
		ctx = cctx.SetTraceIDWorkflowContext(ctx, sc.TraceID)
	}
	return ctx
}
