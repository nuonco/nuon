package cctx

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func SetFlowWorkflowIDContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keys.FlowWorkflowIDCtxKey, id)
}

func SetFlowInstallIDContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keys.FlowInstallIDCtxKey, id)
}
