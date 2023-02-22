package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"github.com/powertoolsdev/api/internal/workflows"
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
	log         *zap.Logger
	repo        repos.AppRepo
	workflowMgr workflows.AppWorkflowManager
}

func NewAppService(db *gorm.DB, tc tclient.Client, log *zap.Logger) *appService {
	repo := repos.NewAppRepo(db)
	return &appService{
		log:         log,
		repo:        repo,
		workflowMgr: workflows.NewAppWorkflowManager(tc),
	}
}

// GetApp: return an app by ID
func (a *appService) GetApp(ctx context.Context, inputID string) (*models.App, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	appID, _ := uuid.Parse(inputID)
	app, err := a.repo.Get(ctx, appID)
	if err != nil {
		a.log.Error("failed to retrieve app",
			zap.String("appID", appID.String()),
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
	if input.GithubInstallID != nil {
		app.GithubInstallID = *input.GithubInstallID
	}

	finalApp, err := a.repo.Update(ctx, app)
	if err != nil {
		a.log.Error("failed to update app",
			zap.Any("app", app),
			zap.String("error", err.Error()))
		return nil, err
	}

	if err := a.workflowMgr.Provision(ctx, app); err != nil {
		a.log.Error("failed to provision app",
			zap.Any("finalApp", *finalApp),
			zap.String("error", err.Error()))
		return nil, err
	}

	return finalApp, nil
}

func (a *appService) UpsertApp(ctx context.Context, input models.AppInput) (*models.App, error) {
	if input.ID != nil {
		return a.updateApp(ctx, input)
	}

	var app models.App
	app.Name = input.Name
	if input.GithubInstallID != nil {
		app.GithubInstallID = *input.GithubInstallID
	}
	app.CreatedByID = *input.CreatedByID

	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	orgID, _ := uuid.Parse(input.OrgID)
	app.OrgID = orgID

	finalApp, err := a.repo.Create(ctx, &app)
	if err != nil {
		a.log.Error("failed to insert app",
			zap.Any("app", app),
			zap.String("error", err.Error()))
		return nil, err
	}

	if err := a.workflowMgr.Provision(ctx, finalApp); err != nil {
		a.log.Error("failed to provision app",
			zap.Any("finalApp", *finalApp),
			zap.String("error", err.Error()))
		return nil, err
	}

	return finalApp, nil
}

func (a *appService) DeleteApp(ctx context.Context, inputID string) (bool, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	appID, _ := uuid.Parse(inputID)
	deleted, err := a.repo.Delete(ctx, appID)
	if err != nil {
		a.log.Error("failed to delete app",
			zap.String("appID", appID.String()),
			zap.String("error", err.Error()))
		return false, err
	}
	return deleted, nil
}

func (a appService) GetOrgApps(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.App, *utils.Page, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	orgID, _ := uuid.Parse(inputID)
	app, pg, err := a.repo.GetPageByOrg(ctx, orgID, options)
	if err != nil {
		a.log.Error("failed to retrieve org's apps",
			zap.String("orgID", orgID.String()),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}
	return app, pg, err
}
