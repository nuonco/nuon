package emitter

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// repairWorkflowContext restores account/org from the emitter row when the
// propagated header carried none. Child starts and continue-as-new re-inject
// headers from this ctx, so an empty header would otherwise persist forever.
func repairWorkflowContext(ctx workflow.Context, emitter *app.QueueEmitter) workflow.Context {
	if emitter == nil {
		return ctx
	}
	if acct, _ := cctx.AccountIDFromContext(ctx); acct == "" && emitter.CreatedByID != "" {
		ctx = cctx.SetAccountIDWorkflowContext(ctx, emitter.CreatedByID)
	}
	if org, _ := cctx.OrgIDFromContext(ctx); org == "" && emitter.OrgID != "" {
		ctx = cctx.SetOrgIDWorkflowContext(ctx, emitter.OrgID)
	}
	return ctx
}
