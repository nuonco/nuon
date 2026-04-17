package jobloop

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/jobs"
)

func (j *jobLoop) executeOutputsJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	outputs, err := handler.Outputs(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get outputs")
	}

	_, err = j.apiClient.CreateJobExecutionOutputs(ctx, job.ID, jobExecution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{
		Outputs: outputs,
	})
	if err != nil {
		return errors.Wrap(err, "unable to write outputs to api")
	}

	return nil
}
