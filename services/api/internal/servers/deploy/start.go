package deploy

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/api/deploy/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (s *server) StartDeploy(
	ctx context.Context,
	req *connect.Request[deployv1.StartDeployRequest],
) (*connect.Response[deployv1.StartDeployResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	deploy := &models.Deploy{
		BuildID:   req.Msg.BuildId,
		InstallID: req.Msg.InstallId,
	}

	err := deploy.NewID()
	if err != nil {
		return nil, err
	}

	deploy, err = s.repo.Create(ctx, deploy)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&deployv1.StartDeployResponse{
		Deploy: deploy.ToProto(),
	}), nil
}
