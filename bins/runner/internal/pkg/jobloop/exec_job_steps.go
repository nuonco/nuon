package jobloop

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) getJobSteps(ctx context.Context, handler jobs.JobHandler) ([]*executeJobStep, error) {
	return []*executeJobStep{
		// validate step
		{
			name:        "fetching",
			fn:          j.executeFetchJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		// validate step
		{
			name:        "validate",
			fn:          j.executeValidateJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		// initialize step
		{
			name:        "initialize",
			fn:          j.executeInitializeJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		// execute step
		{
			name:        "execute",
			fn:          j.executeExecuteJobStep,
			cleanupFn:   j.cleanupJobStep,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInDashProgress,
		},
		// update clean up
		{
			name:        "cleanup",
			fn:          j.executeCleanupJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusCleaningDashUp,
		},
	}, nil
}
