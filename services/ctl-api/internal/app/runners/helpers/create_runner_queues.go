package helpers

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// healthcheckSignalType mirrors the constant in runners/signals/v2/healthcheck
// (defined here to avoid an import cycle through worker/activities).
const healthcheckSignalType queuesignal.SignalType = "healthcheck"

// healthcheckSignalTemplate is a minimal JSON-compatible struct whose shape
// matches healthcheck.Signal. The emitter deserializes this template when
// firing the signal; only the exported fields matter.
type healthcheckSignalTemplate struct {
	RunnerID string `json:"runner_id"`
}

func (s *healthcheckSignalTemplate) Type() queuesignal.SignalType      { return healthcheckSignalType }
func (s *healthcheckSignalTemplate) Validate(_ workflow.Context) error { return nil }
func (s *healthcheckSignalTemplate) Execute(_ workflow.Context) error  { return nil }

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

// CreateRunnerQueues creates one queue per job group for the given runner and registers
// a cron health check emitter. Only called when the parallel-runner-jobs feature flag is enabled.
func (h *Helpers) CreateRunnerQueues(ctx context.Context, runner *app.Runner, settings *app.RunnerGroupSettings) error {
	var healthCheckQueueID string

	for _, group := range allRunnerJobGroups {
		q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     runner.ID,
			OwnerType:   "runners",
			Namespace:   signals.TemporalNamespace,
			Name:        string(group),
			MaxInFlight: settings.MaxInFlightForGroup(group),
			MaxDepth:    100,
		})
		if err != nil {
			return fmt.Errorf("unable to create queue for job group %s: %w", group, err)
		}

		if group == app.RunnerJobGroupHealthChecks {
			healthCheckQueueID = q.ID
		}
	}

	if healthCheckQueueID == "" {
		return fmt.Errorf("health check queue was not created")
	}

	// Create a cron emitter on the health-check queue to drive runner health monitoring.
	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:        healthCheckQueueID,
		Name:           fmt.Sprintf("runner-%s-health-check", runner.ID),
		Description:    "Periodic runner health check",
		Mode:           app.QueueEmitterModeCron,
		CronSchedule:   "* * * * *",
		JitterWindow:   time.Minute,
		SignalType:     healthcheckSignalType,
		SignalTemplate: &healthcheckSignalTemplate{RunnerID: runner.ID},
	}); err != nil {
		return fmt.Errorf("unable to create health check emitter: %w", err)
	}

	return nil
}
