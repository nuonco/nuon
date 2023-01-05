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

type DeploymentService struct {
	repo          repos.DeploymentRepo
	componentRepo repos.ComponentRepo
	githubRepo    repos.GithubRepo
	wkflowMgr     workflows.DeploymentWorkflowManager
}

func NewDeploymentService(db *gorm.DB, temporalClient tclient.Client, tsprt *gh.AppsTransport) *DeploymentService {
	return &DeploymentService{
		repo:          repos.NewDeploymentRepo(db),
		wkflowMgr:     workflows.NewDeploymentWorkflowManager(temporalClient),
		componentRepo: repos.NewComponentRepo(db),
		githubRepo:    repos.NewGithubRepo(tsprt, &zap.Logger{}, &http.Client{}),
	}
}

func (i *DeploymentService) GetDeployment(ctx context.Context, inputID string) (*models.Deployment, error) {
	deploymentID, err := parseID(inputID)
	if err != nil {
		return nil, err
	}

	return i.repo.Get(ctx, deploymentID)
}

func (i *DeploymentService) GetComponentDeployments(ctx context.Context, ids []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
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

func (i *DeploymentService) startDeployment(ctx context.Context, deployment *models.Deployment) error {
	err := i.wkflowMgr.Start(ctx, deployment)
	if err != nil {
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}

func (i *DeploymentService) CreateDeployment(ctx context.Context, componentID string) (*models.Deployment, error) {
	id, err := parseID(componentID)
	if err != nil {
		return nil, err
	}

	_, err = i.componentRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.New("component not found")
	}

	deployment := &models.Deployment{
		ComponentID: id,
	}
	deployment, err = i.repo.Create(ctx, deployment)
	if err != nil {
		return nil, err
	}

	// TODO: distinguish between ComponentType and execute different flows
	ghInstallID, err := strconv.ParseInt(deployment.Component.App.GithubInstallID, 10, 64)
	if err != nil {
		return nil, err
	}

	//TODO: replace _ with token when ready to pass it to the temporal workflow
	_, err = repos.GithubRepo.GetInstallToken(i.githubRepo, ctx, ghInstallID)
	if err != nil {
		return nil, err
	}

	commit, err := repos.GithubRepo.GetCommit(i.githubRepo,
		ctx,
		ghInstallID,
		deployment.Component.GithubRepoOwner,
		deployment.Component.GithubRepo,
		deployment.Component.GithubBranch)
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

	// start deployment workflow
	err = i.startDeployment(ctx, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
