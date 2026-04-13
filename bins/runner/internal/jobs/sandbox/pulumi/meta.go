package pulumi

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

const appRunnerJobTypeSandboxPulumi = models.AppRunnerJobType("sandbox-pulumi")

func (h *handler) Name() string {
	return "sandbox-pulumi"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return appRunnerJobTypeSandboxPulumi
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
