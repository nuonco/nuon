package imagemetadata

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "fetch-image-metadata"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobType("fetch-image-metadata")
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
