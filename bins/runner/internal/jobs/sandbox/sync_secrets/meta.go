package terraform

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "sandbox-sync-secrets"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeSandboxDashSyncDashSecrets
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
