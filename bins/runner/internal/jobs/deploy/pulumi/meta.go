package pulumi

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "pulumi-deploy"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypePulumiDashDeploy
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
