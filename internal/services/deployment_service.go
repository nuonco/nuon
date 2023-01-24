package services

import (
	"context"
	"errors"
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

	// GetDeployment returns deployment for any of the provided component
	// ids
	GetComponentDeployments(context.Context, []string, *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error)
	// CreateDeployment creates a new deployment for the specified component id
	CreateDeployment(context.Context, *models.DeploymentInput) (*models.Deployment, error)
}

type deploymentService struct {
	repo          repos.DeploymentRepo
	componentRepo repos.ComponentRepo
	githubRepo    repos.GithubRepo
	wkflowMgr     workflows.DeploymentWorkflowManager
}

var _ DeploymentService = (*deploymentService)(nil)

func NewDeploymentService(db *gorm.DB, temporalClient tclient.Client, tsprt *gh.AppsTransport) *deploymentService {
	return &deploymentService{
		repo:          repos.NewDeploymentRepo(db),
		wkflowMgr:     workflows.NewDeploymentWorkflowManager(temporalClient),
		componentRepo: repos.NewComponentRepo(db),
		githubRepo:    repos.NewGithubRepo(tsprt, &zap.Logger{}, &http.Client{}),
	}
}

func (i *deploymentService) GetDeployment(ctx context.Context, inputID string) (*models.Deployment, error) {
	deploymentID, err := parseID(inputID)
	if err != nil {
		return nil, err
	}

	return i.repo.Get(ctx, deploymentID)
}

func (i *deploymentService) GetComponentDeployments(ctx context.Context, ids []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	uuids := make([]uuid.UUID, 0)
	for _, v := range ids {
		componentID, err := parseID(v)
		if err != nil {
			return nil, nil, err
		}
		uuids = append(uuids, componentID)
	}

	return i.repo.ListByComponents(ctx, uuids, options)
}

func (i *deploymentService) startDeployment(ctx context.Context, deployment *models.Deployment) error {
	err := i.wkflowMgr.Start(ctx, deployment)
	if err != nil {
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}

func (i *deploymentService) processGithubDeployment(ctx context.Context, deployment *models.Deployment) (*models.Deployment, error) {
	ghInstallID, err := strconv.ParseInt(deployment.Component.App.GithubInstallID, 10, 64)
	if err != nil {
		return nil, err
	}

	// get installation token from github
	//TODO: remove since the workers are getting their own tokens?
	_, err = repos.GithubRepo.GetInstallToken(i.githubRepo, ctx, ghInstallID)
	if err != nil {
		return nil, err
	}

	// get commit info from github
	commit, err := repos.GithubRepo.GetCommit(i.githubRepo,
		ctx,
		ghInstallID,
		deployment.Component.GithubConfig.RepoOwner,
		deployment.Component.GithubConfig.Repo,
		deployment.Component.GithubConfig.Branch)
	if err != nil {
		return nil, err
	}

	// update deployment with commit info
	deployment.CommitAuthor = *commit.Author.Login
	deployment.CommitHash = commit.GetSHA()
	deployment, err = i.repo.Update(ctx, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func (i *deploymentService) CreateDeployment(ctx context.Context, input *models.DeploymentInput) (*models.Deployment, error) {
	id, err := parseID(input.ComponentID)
	if err != nil {
		return nil, err
	}

	_, err = i.componentRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.New("component not found")
	}

	deployment := &models.Deployment{
		ComponentID: id,
		CreatedByID: *input.CreatedByID,
	}
	deployment, err = i.repo.Create(ctx, deployment)
	if err != nil {
		return nil, err
	}

	// do github stuff only when ComponentType is GITHUB_REPO
	if deployment.Component.Type == "GITHUB_REPO" {
		deployment, err = i.processGithubDeployment(ctx, deployment)
		if err != nil {
			return nil, err
		}
	}

	// start deployment workflow
	err = i.startDeployment(ctx, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
