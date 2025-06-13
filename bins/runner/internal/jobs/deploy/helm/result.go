package helm

import (
	"github.com/nuonco/nuon-runner-go/models"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// TODO(jm): pull out the helm resources and their statuses from the release, and write them to the api
func (h *handler) createAPIResult(rel *release.Release, plan string) (*models.ServiceCreateRunnerJobExecutionResultRequest, error) {
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}
	if plan != "" {
		req.Contents = plan
	}

	return req, nil

}
