package jobloop

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) sandboxOutputs() map[string]interface{} {
	return map[string]interface{}{
		"sandbox-outputs": map[string]interface{}{
			"sandbox-mode": true,
			"map": map[string]interface{}{
				"k": "v",
			},
		},
		"image": map[string]interface{}{
			"tag": "local",
		},
	}
}

func (j *jobLoop) executeOutputsJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	var (
		outputs map[string]interface{}
		err     error
	)

	if j.isSandbox(job) {
		outputs = j.sandboxOutputs()
	} else {
		outputs, err = handler.Outputs(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to get outputs")
		}
	}

	_, err = j.apiClient.CreateJobExecutionOutputs(ctx, job.ID, jobExecution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{
		Outputs: outputs,
	})
	if err != nil {
		return errors.Wrap(err, "unable to write outputs to api")
	}

	return nil
}
