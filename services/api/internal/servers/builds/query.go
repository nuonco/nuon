package builds

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	buildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (s *server) QueryBuilds(
	ctx context.Context,
	req *connect.Request[buildv1.QueryBuildsRequest],
) (*connect.Response[buildv1.QueryBuildsResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// use temporal client to get build by ID
	fmt.Println("GET BUILD")
	buildModels := []models.Build{}
	builds := []*buildv1.Build{}
	for _, build := range buildModels {
		builds = append(builds, build.ToProto())
	}

	return connect.NewResponse(&buildv1.QueryBuildsResponse{
		Builds: builds,
	}), nil
}
