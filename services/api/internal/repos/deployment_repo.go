package repos

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_deployment_repo.go -source=deployment_repo.go -package=repos
type DeploymentRepo interface {
	Update(context.Context, *models.Deployment) (*models.Deployment, error)
	Get(context.Context, string) (*models.Deployment, error)
	ListByApps(context.Context, []string, *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error)
	ListByComponents(context.Context, []string, *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error)
	ListByInstalls(context.Context, []string, *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error)
	Create(context.Context, *models.Deployment) (*models.Deployment, error)
}

var _ DeploymentRepo = (*deploymentRepo)(nil)

func NewDeploymentRepo(db *gorm.DB) deploymentRepo {
	return deploymentRepo{
		db: db,
	}
}

type deploymentRepo struct {
	db *gorm.DB
}

func (i deploymentRepo) Get(ctx context.Context, id string) (*models.Deployment, error) {
	var deployment models.Deployment
	if err := i.db.WithContext(ctx).
		Preload("Component.App").
		Preload("Component.App.Installs").
		Preload(clause.Associations).
		First(&deployment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (i deploymentRepo) ListByComponents(ctx context.Context, componentIDs []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	var deployments []*models.Deployment
	tx := i.db.WithContext(ctx).
		Where("component_id IN ?", componentIDs).
		Find(&deployments)
	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&deployments); err != nil {
		return nil, nil, err
	}

	return deployments, &page, nil
}

func (i deploymentRepo) ListByApps(ctx context.Context, appIDs []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	var deployments []*models.Deployment

	tx := i.db.WithContext(ctx).
		Where("component_id IN (?)", i.db.Table("components").
			Select("id").
			Where("app_id IN ?", appIDs)).
		Find(&deployments)

	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&deployments); err != nil {
		return nil, nil, err
	}

	return deployments, &page, nil
}

func (i deploymentRepo) ListByInstalls(ctx context.Context, installIDs []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	var deployments []*models.Deployment

	tx := i.db.WithContext(ctx).
		Where("component_id IN (?)", i.db.Table("components").
			Select("id").
			Where("app_id IN (?)", i.db.Table("installs").
				Select("app_id").
				Where("id IN ?", installIDs))).
		Find(&deployments)

	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&deployments); err != nil {
		return nil, nil, err
	}

	return deployments, &page, nil
}

func (i deploymentRepo) Create(ctx context.Context, deployment *models.Deployment) (*models.Deployment, error) {
	if err := i.db.WithContext(ctx).Create(deployment).Error; err != nil {
		return nil, err
	}

	return i.Get(ctx, deployment.ID)
}

func (i deploymentRepo) Update(ctx context.Context, deployment *models.Deployment) (*models.Deployment, error) {
	if err := i.db.WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Updates(deployment).Error; err != nil {
		return nil, err
	}

	return i.Get(ctx, deployment.ID)
}
