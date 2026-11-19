package containerimage

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "oci-sync"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeOciDashSync
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
