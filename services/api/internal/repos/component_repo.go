package repos

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_component_repo.go -source=component_repo.go -package=repos
type ComponentRepo interface {
	Get(context.Context, string) (*models.Component, error)
	ListByApp(context.Context, string, *models.ConnectionOptions) ([]*models.Component, *utils.Page, error)
	Create(context.Context, *models.Component) (*models.Component, error)
	Update(context.Context, *models.Component) (*models.Component, error)
	Delete(context.Context, string) (bool, error)
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

func (i componentRepo) Get(ctx context.Context, componentID string) (*models.Component, error) {
	component := models.Component{Model: models.Model{ID: componentID}}

	if err := i.db.WithContext(ctx).
		Preload("App.Org").
		Preload(clause.Associations).
		First(&component).Error; err != nil {
		return nil, err
	}

	return &component, nil
}

func (i componentRepo) ListByApp(ctx context.Context, appID string, options *models.ConnectionOptions) ([]*models.Component, *utils.Page, error) {
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

func (i componentRepo) Delete(ctx context.Context, componentID string) (bool, error) {
	var component models.Component
	if err := i.db.WithContext(ctx).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Delete(&component, "id = ?", componentID).Error; err != nil {
		return false, err
	}
	return component.ID != "", nil
}

func (i componentRepo) Create(ctx context.Context, component *models.Component) (*models.Component, error) {
	if err := i.db.WithContext(ctx).Create(component).Error; err != nil {
		return nil, err
	}

	return component, nil
}

func (i componentRepo) Update(ctx context.Context, component *models.Component) (*models.Component, error) {
	if err := i.db.WithContext(ctx).
		Omit(clause.Associations).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Updates(component).Error; err != nil {
		return nil, err
	}

	return component, nil
}
