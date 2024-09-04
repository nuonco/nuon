package terraform

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "sandbox-terraform"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeSandboxDashTerraform
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
