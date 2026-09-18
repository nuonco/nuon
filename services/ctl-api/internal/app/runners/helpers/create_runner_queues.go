package helpers

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// CreateRunnerQueues creates one queue per job group for the given runner.
// Health monitoring is handled by process_healthcheck emitters on per-process queues,
// so no runner-level healthcheck emitter is created here.
func (h *Helpers) CreateRunnerQueues(ctx context.Context, runner *app.Runner, settings *app.RunnerGroupSettings) error {
	return h.EnsureRunnerJobGroupQueues(ctx, runner, settings)
}
