package helm

import (
	"github.com/nuonco/nuon-runner-go/models"
	"helm.sh/helm/v3/pkg/release"
)

// TODO(jm): pull out the helm resources and their statuses from the release, and write them to the api
func (h *handler) createAPIResult(rel *release.Release) (*models.ServiceCreateRunnerJobExecutionResultRequest, error) {
	return &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}, nil
}
