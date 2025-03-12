package check

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "health-check"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeHealthDashCheck
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
