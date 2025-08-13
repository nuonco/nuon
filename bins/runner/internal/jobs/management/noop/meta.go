package noop

import (
	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Name() string {
	return "noop"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeNoop
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
