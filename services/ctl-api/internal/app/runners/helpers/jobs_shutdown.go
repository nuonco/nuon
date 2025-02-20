package helpers

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) CreateShutdownJob(ctx context.Context,
	runnerID string,
	ownerID string,
	logStreamID string,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	return h.CreateRunnerJob(
		ctx,
		runnerID,
		"runners",
		ownerID,
		app.RunnerJobTypeShutDown,
		app.RunnerJobOperationTypeExec,
		logStreamID,
		metadata,
	)
}
