package deploy

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/api/deploy/v1"
)

func (s *server) GetDeploy(
	ctx context.Context,
	req *connect.Request[deployv1.GetDeployRequest],
) (*connect.Response[deployv1.GetDeployResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	deploy, err := s.repo.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&deployv1.GetDeployResponse{
		Deploy: deploy.ToProto(),
	}), nil
}
