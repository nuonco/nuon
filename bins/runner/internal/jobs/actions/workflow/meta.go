package workflow

import (
	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Name() string {
	return "actions-workflow"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeActionsDashWorkflow
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
