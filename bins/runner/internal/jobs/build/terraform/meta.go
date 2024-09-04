package terraform

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "terraform-build"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeTerraformDashModuleDashBuild
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
