package restart

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "restart"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeMngDashRunnerDashRestart
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
