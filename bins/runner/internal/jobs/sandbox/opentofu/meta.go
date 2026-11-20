package opentofu

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "sandbox-opentofu"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeSandboxDashOpentofu
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
