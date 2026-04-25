package handler

import (
	"context"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// UpdateWithStartOptions configures the update sent to a handler workflow.
type UpdateWithStartOptions struct {
	UpdateName   string
	WaitForStage tclient.WorkflowUpdateStage
	Args         []any
}

// UpdateWithStart sends a Temporal update-with-start to the handler workflow
// for the given QueueSignal. Returns the WorkflowUpdateHandle for callers
// to retrieve results via handle.Get().
func UpdateWithStart(
	ctx context.Context,
	tc temporalclient.Client,
	qs *app.QueueSignal,
	opts UpdateWithStartOptions,
) (tclient.WorkflowUpdateHandle, error) {
	startOp := tc.NewWithStartWorkflowOperation(
		tclient.StartWorkflowOptions{
			ID:                       qs.Workflow.ID,
			TaskQueue:                "api",
			WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 0,
			},
		},
		"Handler",
		HandlerRequest{
			QueueID:       qs.QueueID,
			QueueSignalID: qs.ID,
		},
	)
	return tc.UpdateWithStartWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWithStartWorkflowOptions{
			UpdateOptions: tclient.UpdateWorkflowOptions{
				WorkflowID:   qs.Workflow.ID,
				UpdateName:   opts.UpdateName,
				WaitForStage: opts.WaitForStage,
				Args:         opts.Args,
			},
			StartWorkflowOperation: startOp,
		})
}
