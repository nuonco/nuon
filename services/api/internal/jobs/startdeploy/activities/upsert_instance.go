package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
)

type UpsertInstanceResponse struct {
	InstanceID string
}

func (a *activities) UpsertInstanceJob(ctx context.Context, deployID string) (*UpsertInstanceResponse, error) {
	deploy, err := a.deployRepo.Get(ctx, deployID)
	if err != nil {
		return nil, err
	}

	instance := &models.Instance{
		BuildID:     deploy.BuildID,
		DeployID:    deploy.ID,
		InstallID:   deploy.InstallID,
		ComponentID: deploy.Build.ComponentID,
	}

	err = instance.NewID()
	if err != nil {
		return nil, err
	}

	_, err = a.instanceRepo.Create(ctx, []*models.Instance{instance})
	if err != nil {
		return nil, err
	}
	return &UpsertInstanceResponse{}, nil
}
