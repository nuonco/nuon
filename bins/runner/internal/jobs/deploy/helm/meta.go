package helm

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "helm-deploy"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeHelmDashChartDashDeploy
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
