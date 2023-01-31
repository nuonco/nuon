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
//go:generate mockgen -destination=mock_component_repo.go -source=component_repo.go -package=repos
type ComponentRepo interface {
	Get(context.Context, uuid.UUID) (*models.Component, error)
	ListByApp(context.Context, uuid.UUID, *models.ConnectionOptions) ([]*models.Component, *utils.Page, error)
	Create(context.Context, *models.Component) (*models.Component, error)
	Update(context.Context, *models.Component) (*models.Component, error)
	Delete(context.Context, uuid.UUID) (bool, error)
}

var _ ComponentRepo = (*componentRepo)(nil)

func NewComponentRepo(db *gorm.DB) componentRepo {
	return componentRepo{
		db: db,
	}
}

type componentRepo struct {
	db *gorm.DB
}

func (i componentRepo) Get(ctx context.Context, componentID uuid.UUID) (*models.Component, error) {
	var component models.Component
	if err := i.db.WithContext(ctx).
		Preload(clause.Associations).
		First(&component, "id = ?", componentID).Error; err != nil {
		return nil, err
	}
	return &component, nil
}

func (i componentRepo) ListByApp(ctx context.Context, appID uuid.UUID, options *models.ConnectionOptions) ([]*models.Component, *utils.Page, error) {
	var components []*models.Component
	tx := i.db.WithContext(ctx).
		Preload(clause.Associations).
		Where("app_id = ?", appID).
		Find(&components)
	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&components); err != nil {
		return nil, nil, err
	}

	return components, &page, nil
}

func (i componentRepo) Delete(ctx context.Context, componentID uuid.UUID) (bool, error) {
	var component models.Component
	if err := i.db.WithContext(ctx).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Delete(&component, "id = ?", componentID).Error; err != nil {
		return false, err
	}
	return component.ID != uuid.Nil, nil
}

func (i componentRepo) Create(ctx context.Context, component *models.Component) (*models.Component, error) {
	if err := i.db.WithContext(ctx).Create(component).Error; err != nil {
		return nil, err
	}

	return component, nil
}

func (i componentRepo) Update(ctx context.Context, component *models.Component) (*models.Component, error) {
	if err := i.db.WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Updates(component).Error; err != nil {
		return nil, err
	}

	return component, nil
}
