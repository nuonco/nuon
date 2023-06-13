package repos

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_deploy_repo.go -source=deploy_repo.go -package=repos
type DeployRepo interface {
	Update(context.Context, *models.Deploy) (*models.Deploy, error)
	Get(context.Context, string) (*models.Deploy, error)
	Create(context.Context, *models.Deploy) (*models.Deploy, error)
}

var _ DeployRepo = (*deployRepo)(nil)

func NewDeployRepo(db *gorm.DB) deployRepo {
	return deployRepo{
		db: db,
	}
}

type deployRepo struct {
	db *gorm.DB
}

func (i deployRepo) Get(ctx context.Context, id string) (*models.Deploy, error) {
	var deploy models.Deploy
	if err := i.db.WithContext(ctx).
		Preload(clause.Associations).
		First(&deploy, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &deploy, nil
}

func (i deployRepo) Create(ctx context.Context, deploy *models.Deploy) (*models.Deploy, error) {
	if err := i.db.WithContext(ctx).Create(deploy).Error; err != nil {
		return nil, err
	}

	return i.Get(ctx, deploy.ID)
}

func (i deployRepo) Update(ctx context.Context, deploy *models.Deploy) (*models.Deploy, error) {
	if err := i.db.WithContext(ctx).
		Updates(deploy).Error; err != nil {
		return nil, err
	}

	return i.Get(ctx, deploy.ID)
}
