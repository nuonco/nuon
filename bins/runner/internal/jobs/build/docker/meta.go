package docker

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "docker-build"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeDockerDashBuild
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
