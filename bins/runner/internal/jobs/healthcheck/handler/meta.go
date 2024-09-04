package handler

import "github.com/nuonco/nuon-runner-go/models"

func (n *handler) Name() string {
	return "healthcheck"
}

func (n *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeHealthDashCheck
}

func (n *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
