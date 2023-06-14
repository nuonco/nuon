package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gh "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	componentConfig "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/powertoolsdev/mono/services/api/internal/utils"
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
	repo                   repos.DeploymentRepo
	componentRepo          repos.ComponentRepo
	githubRepo             repos.GithubRepo
	githubAppID            string
	githubAppKeySecretName string
	log                    *zap.Logger
}

var _ DeploymentService = (*deploymentService)(nil)

func NewDeploymentService(db *gorm.DB,
	temporalClient tclient.Client,
	tsprt *gh.AppsTransport,
	githubAppID string,
	githubAppKeySecretName string,
	log *zap.Logger) *deploymentService {
	return &deploymentService{
		repo:                   repos.NewDeploymentRepo(db),
		componentRepo:          repos.NewComponentRepo(db),
		githubRepo:             repos.NewGithubRepo(tsprt, &zap.Logger{}, &http.Client{}),
		githubAppID:            githubAppID,
		githubAppKeySecretName: githubAppKeySecretName,
		log:                    log,
	}
}

func (i *deploymentService) GetDeployment(ctx context.Context, deploymentID string) (*models.Deployment, error) {
	deployment, err := i.repo.Get(ctx, deploymentID)
	if err != nil {
		i.log.Error("failed to retrieve deployment",
			zap.String("deploymentID", deploymentID),
			zap.String("error", err.Error()))
		return nil, err
	}

	return deployment, nil
}

func (i *deploymentService) GetComponentDeployments(ctx context.Context, componentIDs []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	deployments, pg, err := i.repo.ListByComponents(ctx, componentIDs, options)
	if err != nil {
		i.log.Error("failed to retrieve component's deployments",
			zap.Any("componentIDs", componentIDs),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}

	return deployments, pg, nil
}

func (i *deploymentService) GetAppDeployments(ctx context.Context, appIDs []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	deployments, pg, err := i.repo.ListByApps(ctx, appIDs, options)
	if err != nil {
		i.log.Error("failed to retrieve apps's deployments",
			zap.Any("appIDs", appIDs),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}

	return deployments, pg, nil
}

func (i *deploymentService) GetInstallDeployments(ctx context.Context, installIDs []string, options *models.ConnectionOptions) ([]*models.Deployment, *utils.Page, error) {
	deployments, pg, err := i.repo.ListByInstalls(ctx, installIDs, options)
	if err != nil {
		i.log.Error("failed to retrieve install's deployments",
			zap.Any("installIDs", installIDs),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}

	return deployments, pg, nil
}

func (i *deploymentService) getRepoBranch(ctx context.Context, ghRepoOwner string, ghRepoName string, ghInstallID string) (string, error) {
	// get repo info from github
	ghRepo, err := repos.GithubRepo.GetRepo(i.githubRepo, ctx, ghInstallID, ghRepoOwner, ghRepoName)
	if err != nil {
		i.log.Error("failed to get commit from GitHub",
			zap.Any("githubInstallID", ghInstallID),
			zap.String("ghRepoOwner", ghRepoOwner),
			zap.String("ghRepoName", ghRepoName),
			zap.String("error", err.Error()))
		return "", fmt.Errorf("error getting the github commit: %w", err)
	}

	return *ghRepo.DefaultBranch, nil
}

func (i *deploymentService) getCommitHash(ctx context.Context, repoOwner string, repoName string, branch string, ghInstallID string, deployment *models.Deployment) (*models.Deployment, error) {
	// get commit info from github
	commit, err := repos.GithubRepo.GetCommit(i.githubRepo,
		ctx,
		ghInstallID,
		repoOwner,
		repoName,
		branch)
	if err != nil {
		i.log.Error("failed to get commit from GitHub",
			zap.Any("githubInstallID", ghInstallID),
			zap.String("repoOwner", repoOwner),
			zap.String("repoName", repoName),
			zap.String("repoBranch", branch),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("error getting the github commit: %w", err)
	}

	commitAuthor := ""
	if commit.Commit.Author.Name != nil {
		commitAuthor = *commit.Commit.Author.Name
	}

	// update deployment with commit info
	deployment.CommitAuthor = commitAuthor
	deployment.CommitHash = commit.GetSHA()
	deployment, err = i.repo.Update(ctx, deployment)
	if err != nil {
		i.log.Error("failed to update deployment with commit info",
			zap.String("commitAuthor", commitAuthor),
			zap.String("commitHash", commit.GetSHA()),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("error updating the deployment: %w", err)
	}

	return deployment, nil
}

func (i *deploymentService) updateComponentConfig(ctx context.Context, ghInstallID string, isPrivate bool, component *models.Component, deployment *models.Deployment) (*models.Deployment, error) {
	// unmarshal existing JSON
	dbConfig, _ := component.Config.MarshalJSON()
	i.log.Info("updating component configuration",
		zap.String("component configuration", string(dbConfig)))
	componentConfiguration := &componentConfig.Component{}
	if err := protojson.Unmarshal(dbConfig, componentConfiguration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input JSON: %w", err)
	}

	if isPrivate {
		// for private repos set: git_ref, github_app_key_id, github_app_key_secret_name, github_install_id
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GitRef = deployment.CommitHash
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubAppKeyId = i.githubAppID
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubAppKeySecretName = i.githubAppKeySecretName
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubInstallId = ghInstallID
	} else {
		// for public repos set only the git_ref
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetPublicGitConfig().GitRef = deployment.CommitHash
	}

	//marshal JSON
	var updatedConfig []byte
	updatedConfig, err := protojson.Marshal(componentConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated component configuration: %w", err)
	}
	component.Config = updatedConfig

	// update component record with new configuration
	component, err = i.componentRepo.Update(ctx, component)
	if err != nil {
		i.log.Error("failed to update component configuration",
			zap.Any("component", *component),
			zap.String("error", err.Error()))
		return nil, err
	}

	return i.GetDeployment(ctx, deployment.GetID())
}

func (i *deploymentService) processGithubRepo(ctx context.Context, vcsCfg *vcsv1.Config, deployment *models.Deployment, component *models.Component) (*models.Deployment, error) {
	var repo string
	connectedGithubConfig := vcsCfg.GetConnectedGithubConfig()
	githubInstallID := "" // start with empty string which also means public git repo

	// if the component config is for a connected github repo,
	// then we have a githubInstallID and should make github API calls using it
	isPrivate := connectedGithubConfig != nil
	if isPrivate {
		// Use the githubInstallID for github API calls
		repo = connectedGithubConfig.Repo
		githubInstallID = component.App.Org.GithubInstallID
	} else {
		publicGitConfig := vcsCfg.GetPublicGitConfig()
		// TODO: temporary workaround, will refactor after the component config retro
		repo = strings.ReplaceAll(publicGitConfig.Repo, "https://github.com/", "")
		repo = strings.ReplaceAll(repo, ".git", "")
	}

	githubInfo := strings.Split(repo, "/") // githubInfo = [RepoOwner, RepoName]
	i.log.Info("querying github for repo details", zap.String("repo", repo), zap.String("githubInstallID", githubInstallID))
	// get default branch from github
	var defaultBranch string
	defaultBranch, err := i.getRepoBranch(ctx, githubInfo[0], githubInfo[1], githubInstallID)
	if err != nil {
		i.log.Error("failed to get github repo details",
			zap.String("repo", repo),
			zap.Any("deployment", deployment),
			zap.String("error", err.Error()))
		return nil, err
	}

	// get commit hash and update deployment record
	deployment, err = i.getCommitHash(ctx, githubInfo[0], githubInfo[1], defaultBranch, githubInstallID, deployment)
	if err != nil {
		i.log.Error("failed to get github commit hash",
			zap.String("repo", repo),
			zap.String("ghInstallID", githubInstallID),
			zap.Any("deployment", deployment),
			zap.String("error", err.Error()))
		return nil, err
	}

	// update component config with github info
	deployment, err = i.updateComponentConfig(ctx, githubInstallID, isPrivate, component, deployment)
	if err != nil {
		i.log.Error("failed to update component configuration",
			zap.String("ghInstallID", githubInstallID),
			zap.Any("component", component),
			zap.Any("deployment", deployment),
			zap.String("error", err.Error()))
		return nil, err
	}

	return deployment, nil
}

func (i *deploymentService) CreateDeployment(ctx context.Context, input *models.DeploymentInput) (*models.Deployment, error) {
	component, err := i.componentRepo.Get(ctx, input.ComponentID)
	if err != nil {
		i.log.Error("failed to get component", zap.String("componentID", input.ComponentID), zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	deployment := &models.Deployment{
		ComponentID: input.ComponentID,
		CreatedByID: *input.CreatedByID,
	}
	deployment.ID = domains.NewDeploymentID()
	deployment, err = i.repo.Create(ctx, deployment)
	if err != nil {
		i.log.Error("failed to create deployment",
			zap.Any("deployment", deployment),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	// retrieve commits if the component configuration contains github info
	dbConfig := &componentConfig.Component{}
	if err = protojson.Unmarshal([]byte(component.Config.String()), dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DB JSON: %w", err)
	}
	if dbConfig.BuildCfg != nil {
		dockerCfg := dbConfig.BuildCfg.GetDockerCfg()
		if dockerCfg != nil {
			deployment, err = i.processGithubRepo(ctx, dockerCfg.GetVcsCfg(), deployment, component)
			if err != nil {
				return nil, fmt.Errorf("failed to process github repo: %w", err)
			}
		}
	}

	return deployment, nil
}
