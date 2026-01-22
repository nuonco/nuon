package shutdown

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "shutdown"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeShutDashDown
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
