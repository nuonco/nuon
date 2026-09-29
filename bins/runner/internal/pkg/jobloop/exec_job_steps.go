package jobloop

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/runner/jobs"
)

func (j *jobLoop) getJobSteps(ctx context.Context, handler jobs.JobHandler) ([]*executeJobStep, error) {
	return []*executeJobStep{
		{
			name:        "resetting",
			fn:          j.executeResetJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		{
			name:        "fetching",
			fn:          j.executeFetchJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		{
			name:        "validate",
			fn:          j.executeValidateJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		{
			name:        "initialize",
			fn:          j.executeInitializeJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInitializing,
		},
		{
			name:        "execute",
			fn:          j.executeExecuteJobStep,
			cleanupFn:   j.cleanupJobStep,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInDashProgress,
		},
		{
			name:        "outputs",
			fn:          j.executeOutputsJobStep,
			cleanupFn:   j.cleanupJobStep,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusInDashProgress,
		},
		{
			name:        "cleanup",
			fn:          j.executeCleanupJobStep,
			cleanupFn:   nil,
			handler:     handler,
			startStatus: models.AppRunnerJobExecutionStatusCleaningDashUp,
		},
	}, nil
}
