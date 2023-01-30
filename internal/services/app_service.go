package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"github.com/powertoolsdev/api/internal/workflows"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_app_service.go -source=app_service.go -package=services
type AppService interface {
	GetApp(context.Context, string) (*models.App, error)
	GetAppBySlug(context.Context, string) (*models.App, error)
	UpsertApp(context.Context, models.AppInput) (*models.App, error)
	DeleteApp(context.Context, string) (bool, error)
	GetAllApps(context.Context, *models.ConnectionOptions) ([]*models.App, *utils.Page, error)
	GetOrgApps(context.Context, string, *models.ConnectionOptions) ([]*models.App, *utils.Page, error)
}

var _ AppService = (*appService)(nil)

type appService struct {
	repo        repos.AppRepo
	workflowMgr workflows.AppWorkflowManager
}

func NewAppService(db *gorm.DB, tc tclient.Client) *appService {
	repo := repos.NewAppRepo(db)
	return &appService{
		repo:        repo,
		workflowMgr: workflows.NewAppWorkflowManager(tc),
	}
}

// GetApp: return an app by ID
func (a *appService) GetApp(ctx context.Context, inputID string) (*models.App, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	appID, _ := uuid.Parse(inputID)
	return a.repo.Get(ctx, appID)
}

// GetBySlug: return an app by slug
func (a *appService) GetAppBySlug(ctx context.Context, slug string) (*models.App, error) {
	return a.repo.GetBySlug(ctx, slug)
}

// UpsertApp: upsert an application
func (a *appService) UpsertApp(ctx context.Context, input models.AppInput) (*models.App, error) {
	var app models.App
	app.Name = input.Name
	app.Slug = slug.Make(input.Name)
	if input.GithubInstallID != nil {
		app.GithubInstallID = *input.GithubInstallID
	}
	app.CreatedByID = *input.CreatedByID

	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	orgID, _ := uuid.Parse(input.OrgID)
	app.OrgID = orgID

	if input.ID != nil {
		// parsing the uuid while ignoring the error handling since we do this at protobuf level
		appID, _ := uuid.Parse(*input.ID)
		app.ID = appID
	}

	finalApp, err := a.repo.Upsert(ctx, &app)
	if err != nil {
		return nil, err
	}

	if err := a.workflowMgr.Provision(ctx, finalApp); err != nil {
		return nil, err
	}

	return finalApp, nil
}

func (a *appService) DeleteApp(ctx context.Context, inputID string) (bool, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	appID, _ := uuid.Parse(inputID)
	return a.repo.Delete(ctx, appID)
}

func (a appService) GetUserApps(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	panic("NOT IMPLEMENTED")
}

func (a appService) GetAllApps(ctx context.Context, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	return a.repo.GetPageAll(ctx, options)
}

func (a appService) GetOrgApps(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	orgID, _ := uuid.Parse(inputID)
	return a.repo.GetPageByOrg(ctx, orgID, options)
}
