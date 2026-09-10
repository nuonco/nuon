package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

// localHintChecksVersion gates running the queue loop's own queue lookups and
// restart-hint checks as local activities. These run on every hint period, so
// they dominated the loop's activity volume — the restart-hint check alone
// accounted for the single largest share of `queues` reads. Histories written
// before this recorded each one as a remote activity command, so replaying them
// against local activities is nondeterministic.
//
// todo(sk): clean up after terminating old workflows
const localHintChecksVersion = "queue-local-hint-checks-v1"

// useLocalHintChecks reports whether this execution may use the local variants.
// Repeat calls with the same change ID return the cached value, so every call
// site in a single execution agrees.
func useLocalHintChecks(ctx workflow.Context) bool {
	return workflow.GetVersion(ctx, localHintChecksVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
}

func getQueueByID(ctx workflow.Context, queueID string) (*app.Queue, error) {
	if useLocalHintChecks(ctx) {
		return activities.LocalAwaitGetQueueByQueueID(ctx, queueID)
	}
	return activities.AwaitGetQueueByQueueID(ctx, queueID)
}

func queueExistsByID(ctx workflow.Context, queueID string) (bool, error) {
	if useLocalHintChecks(ctx) {
		return activities.LocalAwaitQueueExistsByQueueID(ctx, queueID)
	}
	return activities.AwaitQueueExistsByQueueID(ctx, queueID)
}

func checkRestartHint(ctx workflow.Context, req activities.CheckRestartHintRequest) (bool, error) {
	if useLocalHintChecks(ctx) {
		return activities.LocalAwaitCheckRestartHint(ctx, req)
	}
	return activities.AwaitCheckRestartHint(ctx, req)
}

func clearRestartHint(ctx workflow.Context, req activities.ClearRestartHintRequest) error {
	if useLocalHintChecks(ctx) {
		return activities.LocalAwaitClearRestartHint(ctx, req)
	}
	return activities.AwaitClearRestartHint(ctx, req)
}
