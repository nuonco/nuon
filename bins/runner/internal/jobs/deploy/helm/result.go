package helm

import (
	"github.com/nuonco/nuon-runner-go/models"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// TODO(jm): pull out the helm resources and their statuses from the release, and write them to the api
func (h *handler) createAPIResult(rel *release.Release, contents string, display map[string]interface{}) (*models.ServiceCreateRunnerJobExecutionResultRequest, error) {
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		// JUST FOR NOW: the plan is going into both
		Success: true,
		// NOTE: the plan for the machine goes here. but for helm, there really isn't one.
		Contents: contents,
		// NOTE: the plan for human eyes should go in here.
		ContentsDisplay: display,
	}

	return req, nil
}
