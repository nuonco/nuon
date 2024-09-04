package registry

import (
	"github.com/nuonco/nuon-runner-go/models"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// NOTE(jm): once we build out the "results" API, this will become a more first class function, where we build an actual
// request to represent the image here. For now, it's mainly a translation layer, until we add that.
func ToAPIResult(res *ocispec.Descriptor) *models.ServiceCreateRunnerJobExecutionResultRequest {
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}

	return req
}
