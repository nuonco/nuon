package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	gh "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"go.uber.org/zap"

	"github.com/powertoolsdev/api/internal/utils"
	"github.com/powertoolsdev/api/internal/workflows"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_deployment_service.go -source=deployment_service.go -package=services
type DeploymentService interface {
	// GetDeployment returns a deployment by id
	GetDeployment(context.Context, string) (*models.Deployment, error)
	// GetAppDeployments returns deployment for any of the provided app IDs
	GetAppDeployments(context.Context, []string, *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error)
	// GetComponentDeployments returns deployment for any of the provided component IDs
	GetComponentDeployments(context.Context, []string, *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error)
	// GetInstallDeployments returns deployment for any of the provided install IDs
	GetInstallDeployments(context.Context, []string, *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error)
	// CreateDeployment creates a new deployment for the specified component id
	CreateDeployment(context.Context, *models.DeploymentInput) (*models.Deployment, error)
}

type deploymentService struct {
	repo          repos.DeploymentRepo
	componentRepo repos.ComponentRepo
	githubRepo    repos.GithubRepo
	wkflowMgr     workflows.DeploymentWorkflowManager
	log           *zap.Logger
}

var _ DeploymentService = (*deploymentService)(nil)

func NewDeploymentService(db *gorm.DB, temporalClient tclient.Client, tsprt *gh.AppsTransport, log *zap.Logger) *deploymentService {
	return &deploymentService{
		repo:          repos.NewDeploymentRepo(db),
		wkflowMgr:     workflows.NewDeploymentWorkflowManager(temporalClient),
		componentRepo: repos.NewComponentRepo(db),
		githubRepo:    repos.NewGithubRepo(tsprt, &zap.Logger{}, &http.Client{}),
		log:           log,
	}
}

func (i *deploymentService) GetDeployment(ctx context.Context, inputID string) (*models.Deployment, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	deploymentID, _ := uuid.Parse(inputID)

	deployment, err := i.repo.Get(ctx, deploymentID)
	if err != nil {
		i.log.Error("failed to retrieve deployment",
			zap.String("deploymentID", deploymentID.String()),
			zap.String("error", err.Error()))
		return nil, err
	}

	return deployment, nil
}

func (i *deploymentService) GetComponentDeployments(ctx context.Context, ids []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	uuids := make([]uuid.UUID, 0)
	for _, v := range ids {
		// parsing the uuid while ignoring the error handling since we do this at protobuf level
		componentID, _ := uuid.Parse(v)
		uuids = append(uuids, componentID)
	}

	deployments, pg, err := i.repo.ListByComponents(ctx, uuids, options)
	if err != nil {
		i.log.Error("failed to retrieve component's deployments",
			zap.Any("componentIDs", uuids),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}

	return deployments, pg, nil
}

func (i *deploymentService) GetAppDeployments(ctx context.Context, ids []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	uuids := make([]uuid.UUID, 0)
	for _, v := range ids {
		// parsing the uuid while ignoring the error handling since we do this at protobuf level
		appID, _ := uuid.Parse(v)
		uuids = append(uuids, appID)
	}

	deployments, pg, err := i.repo.ListByApps(ctx, uuids, options)
	if err != nil {
		i.log.Error("failed to retrieve apps's deployments",
			zap.Any("appIDs", uuids),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}

	return deployments, pg, nil
}

func (i *deploymentService) GetInstallDeployments(ctx context.Context, ids []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	uuids := make([]uuid.UUID, 0)
	for _, v := range ids {
		// parsing the uuid while ignoring the error handling since we do this at protobuf level
		installID, _ := uuid.Parse(v)
		uuids = append(uuids, installID)
	}

	deployments, pg, err := i.repo.ListByInstalls(ctx, uuids, options)
	if err != nil {
		i.log.Error("failed to retrieve install's deployments",
			zap.Any("installIDs", uuids),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}

	return deployments, pg, nil
}

func (i *deploymentService) startDeployment(ctx context.Context, deployment *models.Deployment) error {
	err := i.wkflowMgr.Start(ctx, deployment)
	if err != nil {
		i.log.Error("failed to start deployment",
			zap.Any("deployment", deployment),
			zap.String("error", err.Error()))
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}

func (i *deploymentService) processGithubDeployment(ctx context.Context, deployment *models.Deployment) (*models.Deployment, error) {
	ghInstallID, err := strconv.ParseInt(deployment.Component.App.GithubInstallID, 10, 64)
	if err != nil {
		i.log.Error("failed to parse GithubInstallID",
			zap.String("GithubInstallID", deployment.Component.App.GithubInstallID),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("error parsing GithubInstallID: %w", err)
	}

	// get installation token from github
	//TODO: remove since the workers are getting their own tokens?
	_, err = repos.GithubRepo.GetInstallToken(i.githubRepo, ctx, ghInstallID)
	if err != nil {
		i.log.Error("failed to get install token from GitHub",
			zap.Any("githubInstallID", ghInstallID),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("error getting an install token from github: %w", err)
	}

	// get commit info from github
	commit, err := repos.GithubRepo.GetCommit(i.githubRepo,
		ctx,
		ghInstallID,
		deployment.Component.GithubConfig.RepoOwner,
		deployment.Component.GithubConfig.Repo,
		deployment.Component.GithubConfig.Branch)
	if err != nil {
		i.log.Error("failed to get commit from GitHub",
			zap.Any("githubInstallID", ghInstallID),
			zap.String("repoOwner", deployment.Component.GithubConfig.RepoOwner),
			zap.String("repo", deployment.Component.GithubConfig.Repo),
			zap.String("branch", deployment.Component.GithubConfig.Branch),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("error getting the github commit: %w", err)
	}

	// update deployment with commit info
	deployment.CommitAuthor = *commit.Author.Login
	deployment.CommitHash = commit.GetSHA()
	deployment, err = i.repo.Update(ctx, deployment)
	if err != nil {
		i.log.Error("failed to update deployment with commit info",
			zap.String("commitAuthor", *commit.Author.Login),
			zap.String("commitHash", commit.GetSHA()),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("error updating the deployment: %w", err)
	}

	return deployment, nil
}

func (i *deploymentService) CreateDeployment(ctx context.Context, input *models.DeploymentInput) (*models.Deployment, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	componentID, _ := uuid.Parse(input.ComponentID)

	_, err := i.componentRepo.Get(ctx, componentID)
	if err != nil {
		i.log.Error("failed to get component",
			zap.String("componentID", componentID.String()),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	deployment := &models.Deployment{
		ComponentID: componentID,
		CreatedByID: *input.CreatedByID,
	}
	deployment, err = i.repo.Create(ctx, deployment)
	if err != nil {
		i.log.Error("failed to create deployment",
			zap.Any("deployment", deployment),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	// do github stuff only when ComponentType is GITHUB_REPO
	if deployment.Component.Type == "GITHUB_REPO" {
		deployment, err = i.processGithubDeployment(ctx, deployment)
		if err != nil {
			i.log.Error("failed to process github deployment",
				zap.Any("deployment", deployment),
				zap.String("error", err.Error()))
			return nil, err
		}
	}

	// start deployment workflow
	err = i.startDeployment(ctx, deployment)
	if err != nil {
		i.log.Error("failed to start deployment",
			zap.Any("deployment", deployment),
			zap.String("error", err.Error()))
		return nil, err
	}

	return deployment, nil
}
