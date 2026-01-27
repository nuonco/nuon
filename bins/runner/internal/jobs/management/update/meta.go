package update

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "update-version"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeMngDashRunnerDashUpdateDashVersion
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
