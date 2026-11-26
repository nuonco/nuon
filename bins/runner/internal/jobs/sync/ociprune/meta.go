package ociprune

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "oci-prune"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeOciDashPrune
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
