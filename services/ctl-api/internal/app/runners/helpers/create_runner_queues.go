package helpers

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) CreateRunnerQueues(ctx context.Context, runner *app.Runner, settings *app.RunnerGroupSettings) error {
	return h.EnsureRunnerJobGroupQueues(ctx, runner, settings)
}
