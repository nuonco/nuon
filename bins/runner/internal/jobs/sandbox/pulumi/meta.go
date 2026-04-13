package pulumi

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "sandbox-pulumi"
}

func (h *handler) JobType() models.AppRunnerJobType {
	// TODO: add a dedicated sandbox-pulumi job type to the API spec and runner SDK.
	// Using pulumi-deploy as a placeholder until the enum is extended.
	return models.AppRunnerJobTypePulumiDashDeploy
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
