package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// repairActivityContext restores account/org from the emitter row when the
// propagated header carried none, so created_by_id is never null on writes.
func repairActivityContext(ctx context.Context, emitter *app.QueueEmitter) context.Context {
	if acct, _ := cctx.AccountIDFromContext(ctx); acct == "" && emitter.CreatedByID != "" {
		ctx = cctx.SetAccountIDContext(ctx, emitter.CreatedByID)
	}
	if org, _ := cctx.OrgIDFromContext(ctx); org == "" && emitter.OrgID != "" {
		ctx = cctx.SetOrgIDContext(ctx, emitter.OrgID)
	}
	return ctx
}
