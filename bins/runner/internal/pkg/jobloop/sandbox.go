package jobloop

import (
	"github.com/nuonco/nuon-runner-go/models"
)

var ignoreSandboxMode []models.AppRunnerJobType = []models.AppRunnerJobType{
	models.AppRunnerJobTypeActionsDashWorkflow,
}

func (j *jobLoop) isSandbox(job *models.AppRunnerJob) bool {
	if job.Type == models.AppRunnerJobTypeShutDashDown {
		return false
	}

	if job.Type == models.AppRunnerJobTypeActionsDashWorkflow {
		return false
	}

	return j.settings.SandboxMode
}
