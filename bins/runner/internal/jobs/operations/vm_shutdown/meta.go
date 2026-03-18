package vmshutdown

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "vm-shutdown"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeVMDashShutDashDown
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
