package services

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/utils"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_app_service.go -source=app_service.go -package=services
type AppService interface {
	GetApp(context.Context, string) (*models.App, error)
	UpsertApp(context.Context, models.AppInput) (*models.App, error)
	DeleteApp(context.Context, string) (bool, error)
	GetOrgApps(context.Context, string, *models.ConnectionOptions) ([]*models.App, *utils.Page, error)
}

var _ AppService = (*appService)(nil)

type appService struct {
	log     *zap.Logger
	orgRepo repos.OrgRepo
	repo    repos.AppRepo
}

func NewAppService(db *gorm.DB, tc tclient.Client, log *zap.Logger) *appService {
	repo := repos.NewAppRepo(db)
	return &appService{
		log:     log,
		orgRepo: repos.NewOrgRepo(db),
		repo:    repo,
	}
}

// GetApp: return an app by ID
func (a *appService) GetApp(ctx context.Context, appID string) (*models.App, error) {
	app, err := a.repo.Get(ctx, appID)
	if err != nil {
		a.log.Error("failed to retrieve app",
			zap.String("appID", appID),
			zap.String("error", err.Error()))
		return nil, err
	}

	return app, nil
}

func (a *appService) updateApp(ctx context.Context, input models.AppInput) (*models.App, error) {
	app, err := a.GetApp(ctx, *input.ID)
	if err != nil {
		a.log.Error("failed to retrieve app",
			zap.Any("input", input),
			zap.String("error", err.Error()))
		return nil, err
	}

	if input.Name != "" {
		app.Name = input.Name
	}

	finalApp, err := a.repo.Update(ctx, app)
	if err != nil {
		a.log.Error("failed to update app",
			zap.Any("app", app),
			zap.String("error", err.Error()))
		return nil, err
	}

	return finalApp, nil
}

func (a *appService) UpsertApp(ctx context.Context, input models.AppInput) (*models.App, error) {
	// check if org exists
	_, err := a.orgRepo.Get(ctx, input.OrgID)
	if err != nil {
		a.log.Error("failed to get org",
			zap.String("orgID", input.OrgID),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get org: %w", err)
	}

	if input.ID != nil {
		return a.updateApp(ctx, input)
	}

	var app models.App
	app.ID = domains.NewAppID()
	app.Name = input.Name
	app.OrgID = input.OrgID

	if input.CreatedByID != nil {
		app.CreatedByID = *input.CreatedByID
	}

	if input.OverrideID != nil {
		app.ID = *input.OverrideID
	}

	finalApp, err := a.repo.Create(ctx, &app)
	if err != nil {
		a.log.Error("failed to insert app",
			zap.Any("app", app),
			zap.String("error", err.Error()))
		return nil, err
	}

	return finalApp, nil
}

func (a *appService) DeleteApp(ctx context.Context, appID string) (bool, error) {
	deleted, err := a.repo.Delete(ctx, appID)
	if err != nil {
		a.log.Error("failed to delete app",
			zap.String("appID", appID),
			zap.String("error", err.Error()))
		return false, err
	}
	return deleted, nil
}

func (a appService) GetOrgApps(ctx context.Context, orgID string, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	app, pg, err := a.repo.GetPageByOrg(ctx, orgID, options)
	if err != nil {
		a.log.Error("failed to retrieve org's apps",
			zap.String("orgID", orgID),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}
	return app, pg, err
}
