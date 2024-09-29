package terraform

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "terraform-deploy"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeRunnerDashTerraform
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
