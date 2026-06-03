package statusactivities

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
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

// LocalAwaitUpdateQueueSignalStatusV2 updates queue signal status as a local activity.
func LocalAwaitUpdateQueueSignalStatusV2(ctx workflow.Context, input UpdateQueueSignalStatusV2Request) error {
	return workflow.ExecuteLocalActivity(withLocalOpts(ctx),
		(*Activities).UpdateQueueSignalStatusV2, input,
	).Get(ctx, nil)
}
