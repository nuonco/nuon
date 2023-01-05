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

type AppService struct {
	repo        repos.AppRepo
	workflowMgr workflows.AppWorkflowManager
}

func NewAppService(db *gorm.DB, tc tclient.Client) *AppService {
	repo := repos.NewAppRepo(db)
	return &AppService{
		repo:        repo,
		workflowMgr: workflows.NewAppWorkflowManager(tc),
	}
}

// GetApp: return an app by ID
func (a *AppService) GetApp(ctx context.Context, inputID string) (*models.App, error) {
	appID, err := parseID(inputID)
	if err != nil {
		return nil, err
	}

	return a.repo.Get(ctx, appID)
}

// GetBySlug: return an app by slug
func (a *AppService) GetAppBySlug(ctx context.Context, slug string) (*models.App, error) {
	return a.repo.GetBySlug(ctx, slug)
}

// UpsertApp: upsert an application
func (a *AppService) UpsertApp(ctx context.Context, input models.AppInput) (*models.App, error) {
	var app models.App
	app.Name = input.Name
	app.Slug = slug.Make(input.Name)
	if input.GithubInstallID != nil {
		app.GithubInstallID = *input.GithubInstallID
	}

	orgID, err := parseID(input.OrgID)
	if err != nil {
		return nil, err
	}
	app.OrgID = orgID

	if input.ID != nil {
		var ID uuid.UUID
		ID, err = parseID(*input.ID)
		if err != nil {
			return nil, err
		}
		app.ID = ID
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

func (a *AppService) DeleteApp(ctx context.Context, inputID string) (bool, error) {
	appID, err := parseID(inputID)
	if err != nil {
		return false, err
	}
	return a.repo.Delete(ctx, appID)
}

func (a AppService) GetUserApps(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	panic("NOT IMPLEMENTED")
}

func (a AppService) GetAllApps(ctx context.Context, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	return a.repo.GetPageAll(ctx, options)
}

func (a AppService) GetOrgApps(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	orgID, err := parseID(inputID)
	if err != nil {
		return nil, nil, err
	}
	return a.repo.GetPageByOrg(ctx, orgID, options)
}
