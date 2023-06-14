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
		InstallID:   deploy.InstallID,
		ComponentID: deploy.Build.ComponentID,
	}

	instance.NewID()

	instances, err := a.instanceRepo.Create(ctx, []*models.Instance{instance})
	if err != nil {
		return nil, err
	}
	return &UpsertInstanceResponse{
		InstanceID: instances[0].ID,
	}, nil
}
