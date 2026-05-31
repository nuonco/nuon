package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
)

// @temporal-gen-v2 workflow
// @execution-timeout 30m
// @task-timeout 1m
func (w *Workflows) Restart(ctx workflow.Context, sreq signals.RequestSignal) error {
	// Legacy event loop restart removed — event loop system has been removed.
	// Queue-based restarts are handled through v2 signals.
	return nil
}
