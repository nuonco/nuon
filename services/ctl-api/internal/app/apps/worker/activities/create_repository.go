package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/ecrrepository"
)

type CreateAppRepositoryRequest struct {
	AppID string

	CreateResponse *ecrrepository.CreateRepositoryResponse
}

// @temporal-gen activity
func (a *Activities) CreateAppRepository(ctx context.Context, req *CreateAppRepositoryRequest) (*app.AppRepository, error) {
	appRep := app.AppRepository{
		AppID:          req.AppID,
		RegistryID:     req.CreateResponse.RegistryID,
		RepositoryName: req.CreateResponse.RepositoryName,
		RepositoryArn:  req.CreateResponse.RepositoryArn,
		RepositoryURI:  req.CreateResponse.RepositoryURI,
		Region:         req.CreateResponse.Region,
	}

	res := a.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Create(&appRep)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app repo")
	}

	return &appRep, nil
}
