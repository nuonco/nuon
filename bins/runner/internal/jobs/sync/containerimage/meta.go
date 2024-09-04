package containerimage

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "container-image-sync"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeContainerDashImageDashSync
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
