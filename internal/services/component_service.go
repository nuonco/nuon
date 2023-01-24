package services

import (
	"context"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_component_service.go -source=component_service.go -package=services
type ComponentService interface {
	GetComponent(context.Context, string) (*models.Component, error)
	GetAppComponents(context.Context, string, *models.ConnectionOptions) ([]*models.Component, *utils.Page, error)
	UpsertComponent(context.Context, models.ComponentInput) (*models.Component, error)
	DeleteComponent(context.Context, string) (bool, error)
}

type componentService struct {
	repo repos.ComponentRepo
}

var _ ComponentService = (*componentService)(nil)

func NewComponentService(db *gorm.DB) ComponentService {
	componentRepo := repos.NewComponentRepo(db)
	return &componentService{
		repo: componentRepo,
	}
}

func (i *componentService) DeleteComponent(ctx context.Context, inputID string) (bool, error) {
	componentID, err := parseID(inputID)
	if err != nil {
		return false, err
	}

	return i.repo.Delete(ctx, componentID)
}

func (i *componentService) GetComponent(ctx context.Context, inputID string) (*models.Component, error) {
	componentID, err := parseID(inputID)
	if err != nil {
		return nil, err
	}

	return i.repo.Get(ctx, componentID)
}

func (i *componentService) GetAppComponents(ctx context.Context, ID string, options *models.ConnectionOptions) ([]*models.Component, *utils.Page, error) {
	appID, err := parseID(ID)
	if err != nil {
		return nil, nil, err
	}

	return i.repo.ListByApp(ctx, appID, options)
}
func (i *componentService) updateComponent(ctx context.Context, input models.ComponentInput) (*models.Component, error) {
	component, err := i.GetComponent(ctx, *input.ID)
	if err != nil {
		return nil, err
	}

	// NOTE: we do not support changing region or account ID on an install
	component.Name = input.Name
	component.BuildImage = input.BuildImage
	component.Type = string(input.Type)
	if input.GithubRepo != nil {
		component.GithubRepo = *input.GithubRepo
	}
	if input.GithubDir != nil {
		component.GithubDir = *input.GithubDir
	}
	if input.GithubBranch != nil {
		component.GithubBranch = *input.GithubBranch
	}
	if input.GithubRepoOwner != nil {
		component.GithubRepoOwner = *input.GithubRepoOwner
	}
	if input.GithubConfig != nil {
		if component.GithubConfig != nil {
			component.GithubConfig.Repo = input.GithubConfig.Repo
			component.GithubConfig.Directory = *input.GithubConfig.Directory
			component.GithubConfig.RepoOwner = *input.GithubConfig.RepoOwner
			component.GithubConfig.Branch = *input.GithubConfig.Branch
		} else {
			component.GithubConfig = &models.GithubConfig{
				Repo:      input.GithubConfig.Repo,
				Directory: *input.GithubConfig.Directory,
				RepoOwner: *input.GithubConfig.RepoOwner,
				Branch:    *input.GithubConfig.Branch,
			}
		}
		component.VcsConfig = component.GithubConfig
	}

	return i.repo.Update(ctx, component)
}

func (i *componentService) UpsertComponent(ctx context.Context, input models.ComponentInput) (*models.Component, error) {
	if input.ID != nil {
		return i.updateComponent(ctx, input)
	}

	var component models.Component
	component.Name = input.Name
	appID, err := parseID(input.AppID)
	if err != nil {
		return nil, err
	}
	component.AppID = appID
	component.BuildImage = input.BuildImage
	component.Type = string(input.Type)
	component.CreatedByID = input.CreatedByID
	if input.GithubRepo != nil {
		component.GithubRepo = *input.GithubRepo
	}
	if input.GithubDir != nil {
		component.GithubDir = *input.GithubDir
	}
	if input.GithubBranch != nil {
		component.GithubBranch = *input.GithubBranch
	}
	if input.GithubRepoOwner != nil {
		component.GithubRepoOwner = *input.GithubRepoOwner
	}
	if input.GithubConfig != nil {
		component.GithubConfig = &models.GithubConfig{
			Repo:      input.GithubConfig.Repo,
			Directory: *input.GithubConfig.Directory,
			RepoOwner: *input.GithubConfig.RepoOwner,
			Branch:    *input.GithubConfig.Branch,
		}
		component.VcsConfig = component.GithubConfig
	}

	return i.repo.Create(ctx, &component)
}
