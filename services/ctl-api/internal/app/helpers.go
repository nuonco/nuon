package app

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func createdByIDFromTemporalContext(ctx workflow.Context) string {
	val := ctx.Value(keys.AccountIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}
	return valStr
}

func createdByIDFromContext(ctx context.Context) string {
	return keys.CreatedByIDFromContext(ctx)
}

func orgIDFromContext(ctx context.Context) string {
	return keys.OrgIDFromContext(ctx)
}

func flowWorkflowIDFromContext(ctx context.Context) string {
	return keys.FlowWorkflowIDFromContext(ctx)
}

func flowInstallIDFromContext(ctx context.Context) string {
	return keys.FlowInstallIDFromContext(ctx)
}

func logstreamIDFromContext(ctx context.Context) string {
	val := ctx.Value(keys.LogStreamCtxKey)
	valObj, ok := val.(*LogStream)
	if !ok {
		return ""
	}
	return valObj.ID
}
