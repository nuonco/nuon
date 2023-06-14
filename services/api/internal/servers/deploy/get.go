package deploy

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/api/deploy/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
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

func (s *server) GetDeploysByInstance(
	ctx context.Context,
	req *connect.Request[deployv1.GetDeploysByInstanceRequest],
) (*connect.Response[deployv1.GetDeploysByInstanceResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	deploys, _, err := s.repo.ListByInstance(ctx, req.Msg.InstanceId, &models.ConnectionOptions{})
	if err != nil {
		return nil, err
	}

	protos := make([]*deployv1.Deploy, len(deploys))
	for idx, deploy := range deploys {
		protoComponent := deploy.ToProto()
		if err != nil {
			return nil, err
		}
		protos[idx] = protoComponent
	}

	return connect.NewResponse(&deployv1.GetDeploysByInstanceResponse{
		Deploys: protos,
	}), nil
}
