package pulumi

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "pulumi-build"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypePulumiDashBuild
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
