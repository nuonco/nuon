package registry

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ToAPIResult builds the base success request for an image build/copy. The
// caller is expected to populate source-identity fields (SourceRef,
// SourceImage, ResolvedTag, SourceDigest, SourceMediaType, ResolvedAt, NoOp)
// directly on the returned struct using values derived from `res` and the
// resolver outcome.
//
// NOTE(jm): once we build out the "results" API, this will become a more first
// class function, where we build an actual request to represent the image
// here. For now, it's mainly a translation layer, until we add that.
func ToAPIResult(res *ocispec.Descriptor) *models.ServiceCreateRunnerJobExecutionResultRequest {
	_ = res
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}

	return req
}
