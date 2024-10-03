package jobloop

import (
	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
)

func (j *jobLoop) getHandler(job *models.AppRunnerJob) (jobs.JobHandler, error) {
	for _, handler := range j.jobHandlers {
		if err := jobs.Matches(job, handler); err == nil {
			return handler, nil
		}
	}

	return nil, errors.Newf("job handler not found for %s job", job.Type)
}
