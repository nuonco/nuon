package registry

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func ToAPIResult(res *ocispec.Descriptor) *models.ServiceCreateRunnerJobExecutionResultRequest {
	_ = res
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}

	return req
}
