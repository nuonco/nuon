package deploy

import (
	"context"
	"errors"
	"fmt"

	"github.com/bufbuild/connect-go"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/api/deploy/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"gorm.io/gorm"
)

func (s *server) StartDeploy(
	ctx context.Context,
	req *connect.Request[deployv1.StartDeployRequest],
) (*connect.Response[deployv1.StartDeployResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	instance, err := s.instanceRepo.GetByInstallAndComponent(ctx, req.Msg.InstallId, req.Msg.ComponentId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if instance == nil {
		instance = &models.Instance{
			BuildID:     req.Msg.BuildId,
			InstallID:   req.Msg.InstallId,
			ComponentID: req.Msg.ComponentId,
		}
		instance.NewID()
		instances, err := s.instanceRepo.Create(ctx, []*models.Instance{instance})
		if err != nil {
			return nil, err
		}
		instance = instances[0]
	}

	fmt.Printf("instance %+v\n", instance)
	deploy := &models.Deploy{
		BuildID:    req.Msg.BuildId,
		InstallID:  req.Msg.InstallId,
		InstanceID: instance.ID,
	}

	err = deploy.NewID()
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
