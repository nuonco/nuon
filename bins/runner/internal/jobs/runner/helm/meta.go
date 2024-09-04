package helm

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "runner-helm"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeRunnerDashHelm
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
