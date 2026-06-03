package activities

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

var defaultLocalActivityOpts = workflow.LocalActivityOptions{
	ScheduleToCloseTimeout: 10 * time.Second,
	RetryPolicy: &temporal.RetryPolicy{
		MaximumAttempts: 1,
	},
}

func withLocalOpts(ctx workflow.Context) workflow.Context {
	return workflow.WithLocalActivityOptions(ctx, defaultLocalActivityOpts)
}

// LocalAwaitGetQueueSignalByQueueSignalID fetches a queue signal by ID as a local activity.
func LocalAwaitGetQueueSignalByQueueSignalID(ctx workflow.Context, queueSignalID string) (*app.QueueSignal, error) {
	var result *app.QueueSignal
	err := workflow.ExecuteLocalActivity(withLocalOpts(ctx),
		(*Activities).QueueInternalGetQueueSignal,
		GetQueueSignalRequest{QueueSignalID: queueSignalID},
	).Get(ctx, &result)
	return result, err
}

// LocalAwaitGetQueueSignalSignalByQueueSignalID deserializes a signal from the DB as a local activity.
func LocalAwaitGetQueueSignalSignalByQueueSignalID(ctx workflow.Context, queueSignalID string) (signal.Signal, error) {
	var result signal.Signal
	err := workflow.ExecuteLocalActivity(withLocalOpts(ctx),
		(*Activities).QueueInternalGetQueueSignalSignal,
		GetQueueSignalSignalRequest{QueueSignalID: queueSignalID},
	).Get(ctx, &result)
	return result, err
}

// LocalAwaitUpdateQueueSignalRunID persists handler run ID as a local activity.
func LocalAwaitUpdateQueueSignalRunID(ctx workflow.Context, req *UpdateQueueSignalRunIDRequest) error {
	return workflow.ExecuteLocalActivity(withLocalOpts(ctx),
		(*Activities).UpdateQueueSignalRunID, req,
	).Get(ctx, nil)
}

// LocalAwaitIncrementQueueSignalExecutionCount increments execution count as a local activity.
func LocalAwaitIncrementQueueSignalExecutionCount(ctx workflow.Context, req *IncrementQueueSignalExecutionCountRequest) error {
	return workflow.ExecuteLocalActivity(withLocalOpts(ctx),
		(*Activities).IncrementQueueSignalExecutionCount, req,
	).Get(ctx, nil)
}

// LocalAwaitCheckCANRequested checks for CAN hint in queue metadata as a local activity.
func LocalAwaitCheckCANRequested(ctx workflow.Context, req CheckCANRequestedRequest) (bool, error) {
	var result bool
	err := workflow.ExecuteLocalActivity(withLocalOpts(ctx),
		(*Activities).QueueInternalCheckCANRequested, req,
	).Get(ctx, &result)
	return result, err
}

// LocalAwaitClearCANRequested clears CAN hint from queue metadata as a local activity.
func LocalAwaitClearCANRequested(ctx workflow.Context, req ClearCANRequestedRequest) error {
	return workflow.ExecuteLocalActivity(withLocalOpts(ctx),
		(*Activities).QueueInternalClearCANRequested, req,
	).Get(ctx, nil)
}
