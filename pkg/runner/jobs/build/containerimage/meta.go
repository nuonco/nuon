package containerimage

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "container-image-build"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeContainerDashImageDashBuild
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
