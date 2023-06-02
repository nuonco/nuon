package repos

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_app_repo.go -source=app_repo.go -package=repos
type AppRepo interface {
	Get(context.Context, string) (*models.App, error)
	GetPageByOrg(context.Context, string, *models.ConnectionOptions) ([]*models.App, *utils.Page, error)
	Create(context.Context, *models.App) (*models.App, error)
	Update(context.Context, *models.App) (*models.App, error)
	Delete(context.Context, string) (bool, error)
}

var _ AppRepo = (*appRepo)(nil)

type appRepo struct {
	db *gorm.DB
}

func NewAppRepo(db *gorm.DB) appRepo {
	return appRepo{
		db: db,
	}
}

func (a appRepo) Get(ctx context.Context, appID string) (*models.App, error) {
	var app models.App
	if err := a.db.WithContext(ctx).First(&app, "id = ?", appID).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (a appRepo) Create(ctx context.Context, app *models.App) (*models.App, error) {
	if err := a.db.WithContext(ctx).Create(app).Error; err != nil {
		return nil, err
	}

	return app, nil
}

func (a appRepo) Update(ctx context.Context, app *models.App) (*models.App, error) {
	if err := a.db.WithContext(ctx).
		Omit(clause.Associations).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Updates(app).Error; err != nil {
		return nil, err
	}

	return app, nil
}

func (a appRepo) Delete(ctx context.Context, appID string) (bool, error) {
	var app models.App
	err := a.db.WithContext(ctx).Delete(&app, "id = ?", appID).Error
	if err != nil {
		return false, err
	}
	// app.ID will be null if the app was found (and soft-deleted) or not found (non existing or already soft-deleted)
	if app.ID != "" {
		return false, nil
	}

	return true, err
}

func (a appRepo) GetPageByOrg(ctx context.Context, orgID string, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	var apps []*models.App
	tx := a.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&apps)
	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&apps); err != nil {
		return nil, nil, err
	}

	return apps, &page, nil
}
