package handler

import (
	"context"
	"fmt"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	updateWithStartMaxAttempts  = 15
	updateWithStartInitialDelay = 50 * time.Millisecond
	updateWithStartMaxDelay     = 5 * time.Second
)

func RetryUpdate(ctx context.Context, updateName string, retryable func(error) bool, submit func() error) error {
	delay := updateWithStartInitialDelay
	var lastErr error
	for attempt := 1; attempt <= updateWithStartMaxAttempts; attempt++ {
		err := submit()
		if err == nil {
			return nil
		}
		if !retryable(err) {
			return err
		}
		lastErr = err
		if attempt == updateWithStartMaxAttempts {
			break
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		delay = min(delay*2, updateWithStartMaxDelay)
	}
	return fmt.Errorf("%s handler not registered after %d attempts: %w", updateName, updateWithStartMaxAttempts, lastErr)
}

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
	taskQueue := qs.Workflow.TaskQueue
	if taskQueue == "" {
		taskQueue = workflows.APITaskQueue
	}
	startOp := tc.NewWithStartWorkflowOperation(
		tclient.StartWorkflowOptions{
			ID:                       qs.Workflow.ID,
			TaskQueue:                taskQueue,
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
