package builds

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	buildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (s *server) CancelBuild(
	ctx context.Context,
	req *connect.Request[buildv1.CancelBuildRequest],
) (*connect.Response[buildv1.CancelBuildResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// use gorm model to create build record
	// use temporal client to start build workflow
	fmt.Println("CREATE BUILD")
	build := models.Build{}

	return connect.NewResponse(&buildv1.CancelBuildResponse{
		Build: build.ToProto(),
	}), nil
}
