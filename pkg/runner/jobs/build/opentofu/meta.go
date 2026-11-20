package opentofu

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "opentofu-build"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeOpentofuDashModuleDashBuild
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
