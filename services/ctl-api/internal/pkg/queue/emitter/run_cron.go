package emitter

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	cronTickerIDTemplate    = "queue-emitter-cron-%s-%s"
	cronParentRunDuration   = 24 * time.Hour
	cronParentCheckInterval = 5 * time.Minute
)

func (e *emitterWorkflow) runCronMode(ctx workflow.Context, l *zap.Logger, emitter *app.QueueEmitter) (bool, error) {
	l.Info("running in cron mode",
		zap.String("emitter-id", e.emitterID),
		zap.String("queue-id", emitter.QueueID),
		zap.Int64("emit-count", e.state.EmitCount),
	)

	// Start the cron ticker child workflow
	childWorkflowID := fmt.Sprintf(cronTickerIDTemplate, emitter.QueueID, e.emitterID)
	if err := e.ensureCronTickerRunning(ctx, l, emitter, childWorkflowID); err != nil {
		return false, errors.Wrap(err, "failed to ensure cron ticker running")
	}

	// Parent workflow runs for a duration, then continues-as-new to prevent unbounded history
	runUntil := workflow.Now(ctx).Add(cronParentRunDuration)

	for workflow.Now(ctx).Before(runUntil) {
		if e.stopped {
			l.Info("emitter stopped")
			return true, nil
		}

		if err := workflow.Sleep(ctx, cronParentCheckInterval); err != nil {
			return false, err
		}
	}

	l.Info("parent run duration complete, continuing as new")
	return false, nil
}

func (e *emitterWorkflow) ensureCronTickerRunning(ctx workflow.Context, l *zap.Logger, emitter *app.QueueEmitter, childWorkflowID string) error {
	parentInfo := workflow.GetInfo(ctx)

	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:            childWorkflowID,
		TaskQueue:             parentInfo.TaskQueueName,
		CronSchedule:          emitter.CronSchedule,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	})

	req := CronTickerWorkflowRequest{
		QueueID:   emitter.QueueID,
		EmitterID: e.emitterID,
	}

	childFuture := workflow.ExecuteChildWorkflow(childCtx, "CronTicker", req)

	var childExec workflow.Execution
	if err := childFuture.GetChildWorkflowExecution().Get(ctx, &childExec); err != nil {
		return errors.Wrap(err, "failed to start cron ticker child workflow")
	}

	l.Info("cron ticker child workflow running",
		zap.String("child-workflow-id", childExec.ID),
		zap.String("child-run-id", childExec.RunID),
		zap.String("cron-schedule", emitter.CronSchedule),
	)

	return nil
}
