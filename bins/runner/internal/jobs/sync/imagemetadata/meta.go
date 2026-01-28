package imagemetadata

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "fetch-image-metadata"
}

func (h *handler) JobType() models.AppRunnerJobType {
	// Note: Using string literal until nuon-runner-go SDK is updated with this constant
	return models.AppRunnerJobType("fetch-image-metadata")
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
