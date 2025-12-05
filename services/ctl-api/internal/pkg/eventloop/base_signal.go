package eventloop

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/propagator"
)

type BaseSignal struct {
	CtxPayload      *propagator.Payload `json:"ctx_payload"`
	SignalListeners []SignalListener    `json:"signal_listeners"`
	CGroup          string              `json:"cgroup`
}

func (BaseSignal) WorkflowName() string {
	return "EventLoop"
}

func (BaseSignal) WorkflowID(id string) string {
	return "event-loop-" + id
}

func (b BaseSignal) ConcurrencyGroup() string {
	return b.CGroup
}

func (b *BaseSignal) PropagateContext(ctx cctx.ValueContext) error {
	payload, err := propagator.FetchPayload(ctx)
	if err != nil {
		return err
	}

	b.CtxPayload = payload
	return nil
}

func (b *BaseSignal) GetWorkflowContext(ctx workflow.Context) workflow.Context {
	if b.CtxPayload == nil {
		return ctx
	}

	ctx = cctx.SetAccountIDWorkflowContext(ctx, b.CtxPayload.AccountID)
	ctx = cctx.SetOrgIDWorkflowContext(ctx, b.CtxPayload.OrgID)
	ctx = cctx.SetTraceIDWorkflowContext(ctx, b.CtxPayload.TraceID)
	if b.CtxPayload.LogStream != nil {
		ctx = cctx.SetLogStreamWorkflowContext(ctx, b.CtxPayload.LogStream)
	}

	return ctx
}

func (b *BaseSignal) GetContext(ctx context.Context) context.Context {
	if b.CtxPayload == nil {
		return ctx
	}

	ctx = cctx.SetAccountIDContext(ctx, b.CtxPayload.AccountID)
	ctx = cctx.SetOrgIDContext(ctx, b.CtxPayload.OrgID)
	ctx = cctx.SetTraceIDContext(ctx, b.CtxPayload.TraceID)
	if b.CtxPayload.LogStream != nil {
		ctx = cctx.SetLogStreamContext(ctx, b.CtxPayload.LogStream)
	}

	return ctx
}

func (b *BaseSignal) Listeners() []SignalListener {
	return b.SignalListeners
}

func (BaseSignal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get org from context: %w", err)
	}

	return org, nil
}
