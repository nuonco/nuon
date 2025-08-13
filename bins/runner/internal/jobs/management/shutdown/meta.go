package shutdown

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "shutdown"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeMngDashShutDashDown
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
