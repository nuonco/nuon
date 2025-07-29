package kubernetes_manifest

import (
	"github.com/nuonco/nuon-runner-go/models"
	release "helm.sh/helm/v4/pkg/release/v1"
)

func (h *handler) createAPIResult(rel *release.Release, contents string, display map[string]interface{}) (*models.ServiceCreateRunnerJobExecutionResultRequest, error) {
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:         true,
		Contents:        contents,
		ContentsDisplay: display,
	}

	return req, nil
}
