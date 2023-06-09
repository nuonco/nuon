package activities

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"gorm.io/gorm"
)

type activities struct {
	v            *validator.Validate
	db           *gorm.DB
	instanceRepo repos.InstanceRepo
	deployRepo   repos.DeployRepo
}

func NewActivities(v *validator.Validate, db *gorm.DB) *activities {
	return &activities{
		v:            v,
		db:           db,
		instanceRepo: repos.NewInstanceRepo(db),
		deployRepo:   repos.NewDeployRepo(db),
	}
}

type GetIDsResponse struct {
	DeployID  string `faker:"shortID"`
	InstallID string `faker:"shortID"`
	BuildID   string `faker:"shortID"`
}

func (a *activities) GetIDs(ctx context.Context, deployID string) (*GetIDsResponse, error) {
	deploy, err := a.deployRepo.Get(ctx, deployID)
	if err != nil {
		return nil, err
	}

	return &GetIDsResponse{
		DeployID:  deploy.ID,
		InstallID: deploy.InstallID,
		BuildID:   deploy.BuildID,
	}, nil
}
