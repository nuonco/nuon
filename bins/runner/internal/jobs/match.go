package jobs

import (
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
)

func Matches(job *models.AppRunnerJob, handler JobHandler) error {
	if models.AppRunnerJobType(job.Type) != handler.JobType() {
		return fmt.Errorf("invalid job type %s", job.Type)
	}

	//if models.AppRunnerJobStatus(job.Status) != handler.JobStatus() {
	//return fmt.Errorf("invalid job type %s", job.Type)
	//}

	return nil
}
