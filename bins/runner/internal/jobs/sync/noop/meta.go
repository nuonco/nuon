package noop

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "noop-sync"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeNoopDashSync
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
