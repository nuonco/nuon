package kubernetes_manifest

import "github.com/nuonco/nuon-runner-go/models"

func (h *handler) Name() string {
	return "kubernetes-manifest-deploy"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobTypeKubernetesDashManifestDashDeploy
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
