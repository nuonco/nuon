package client

import (
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/awaiter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// AwaitQueueSignal starts an AwaitSignal child workflow that listens for the
// handler's completion signal. The child workflow continues-as-new aggressively
// to keep history small and checks DB on each run as a safety net.
func AwaitQueueSignal(ctx workflow.Context, queueSignalID string) (*handler.FinishedResponse, error) {
	qs, err := AwaitGetQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, err
	}

	timeout := signal.DefaultTimeout
	if qs.Signal.Signal != nil {
		timeout = signal.DeriveTimeout(qs.Signal.Signal)
	}

	return awaiter.AwaitAwaitSignalWorkflow(ctx, awaiter.AwaitSignalWorkflowRequest{
		QueueSignalID: queueSignalID,
		Timeout:       timeout,
	}, &workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: timeout,
		ParentClosePolicy:        enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	})
}
