package kubernetes_manifest

import "github.com/nuonco/nuon-runner-go/models"

// TODO: Once nuon-runner-go is regenerated from the updated ctl-api swagger spec,
// replace this with models.AppRunnerJobTypeKubernetesDashManifestDashBuild
const appRunnerJobTypeKubernetesManifestBuild models.AppRunnerJobType = "kubernetes-manifest-build"

func (h *handler) Name() string {
	return "kubernetes-manifest-build"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return appRunnerJobTypeKubernetesManifestBuild
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
