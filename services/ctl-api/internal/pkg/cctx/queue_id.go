package cctx

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func SetQueueIDContext(ctx context.Context, queueID string) context.Context {
	return context.WithValue(ctx, keys.QueueIDCtxKey, queueID)
}

func SetQueueIDWorkflowContext(ctx workflow.Context, queueID string) workflow.Context {
	return workflow.WithValue(ctx, keys.QueueIDCtxKey, queueID)
}

func QueueIDFromContext(ctx ValueContext) string {
	queueID, _ := ctx.Value(keys.QueueIDCtxKey).(string)
	return queueID
}
