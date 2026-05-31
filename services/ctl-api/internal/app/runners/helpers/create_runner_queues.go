package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

var allRunnerJobGroups = []app.RunnerJobGroup{
	app.RunnerJobGroupHealthChecks,
	app.RunnerJobGroupSync,
	app.RunnerJobGroupBuild,
	app.RunnerJobGroupDeploy,
	app.RunnerJobGroupSandbox,
	app.RunnerJobGroupRunner,
	app.RunnerJobGroupOperations,
	app.RunnerJobGroupManagement,
	app.RunnerJobGroupActions,
}

// CreateRunnerQueues creates one queue per job group for the given runner.
// Health monitoring is handled by process_healthcheck emitters on per-process queues,
// so no runner-level healthcheck emitter is created here.
func (h *Helpers) CreateRunnerQueues(ctx context.Context, runner *app.Runner, settings *app.RunnerGroupSettings) error {
	for _, group := range allRunnerJobGroups {
		if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     runner.ID,
			OwnerType:   "runners",
			Namespace:   signals.TemporalNamespace,
			Name:        string(group),
			MaxInFlight: settings.MaxInFlightForGroup(group),
			MaxDepth:    100,
		}); err != nil {
			return fmt.Errorf("unable to create queue for job group %s: %w", group, err)
		}
	}

	return nil
}
