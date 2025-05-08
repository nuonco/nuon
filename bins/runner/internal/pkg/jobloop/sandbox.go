package jobloop

import (
	"github.com/nuonco/nuon-runner-go/models"
)

func (j *jobLoop) isSandbox(job *models.AppRunnerJob) bool {
	if job.Type == models.AppRunnerJobTypeShutDashDown {
		return false
	}

	return j.settings.SandboxMode
}
