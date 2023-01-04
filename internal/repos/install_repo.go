package repos

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_install_repo.go -source=install_repo.go -package=repos
type InstallRepo interface {
	Get(context.Context, uuid.UUID) (*models.Install, error)
	ListByApp(context.Context, uuid.UUID, *models.ConnectionOptions) ([]*models.Install, *utils.Page, error)
	Create(context.Context, *models.Install) (*models.Install, error)
	Update(context.Context, *models.Install) (*models.Install, error)
	Delete(context.Context, uuid.UUID) (bool, error)
}

var _ InstallRepo = (*installRepo)(nil)

func NewInstallRepo(db *gorm.DB) installRepo {
	return installRepo{
		db: db,
	}
}

type installRepo struct {
	db *gorm.DB
}

func (i installRepo) Get(ctx context.Context, installID uuid.UUID) (*models.Install, error) {
	var install models.Install
	if err := i.db.WithContext(ctx).Preload(clause.Associations).First(&install, "installs.id = ?", installID).Error; err != nil {
		return nil, err
	}
	if install.AWSSettings != nil {
		install.Settings = install.AWSSettings
	}
	return &install, nil
}

func (i installRepo) ListByApp(
	ctx context.Context,
	appID uuid.UUID,
	options *models.ConnectionOptions,
) ([]*models.Install, *utils.Page, error) {
	var installs []*models.Install
	tx := i.db.WithContext(ctx).Where("app_id = ?", appID).Find(&installs)
	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&installs); err != nil {
		return nil, nil, err
	}

	return installs, &page, nil
}

func (i installRepo) Delete(ctx context.Context, installID uuid.UUID) (bool, error) {
	if err := i.db.WithContext(ctx).Model(&models.Install{Model: models.Model{ID: installID}}).Association("Domain").Delete(); err != nil {
		return false, err
	}

	if err := i.db.WithContext(ctx).Model(&models.Install{Model: models.Model{ID: installID}}).Association("AWSSettings").Delete(); err != nil {
		return false, err
	}

	if err := i.db.WithContext(ctx).Model(&models.Install{Model: models.Model{ID: installID}}).Association("GCPSettings").Delete(); err != nil {
		return false, err
	}

	var install models.Install
	if err := i.db.WithContext(ctx).Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Delete(&install, "id = ?", installID).Error; err != nil {
		return false, err
	}
	return install.ID != uuid.Nil, nil
}

func (i installRepo) Create(ctx context.Context, install *models.Install) (*models.Install, error) {
	if err := i.db.WithContext(ctx).Create(install).Error; err != nil {
		return nil, err
	}

	return install, nil
}

func (i installRepo) Update(ctx context.Context, install *models.Install) (*models.Install, error) {
	if err := i.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(install).Error; err != nil {
		return nil, err
	}

	return install, nil
}
