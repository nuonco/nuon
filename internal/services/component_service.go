package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	componentConfig "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/datatypes"
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
	log  *zap.Logger
	repo repos.ComponentRepo
}

var _ ComponentService = (*componentService)(nil)

func NewComponentService(db *gorm.DB, log *zap.Logger) ComponentService {
	componentRepo := repos.NewComponentRepo(db)
	return &componentService{
		log:  log,
		repo: componentRepo,
	}
}

func (i *componentService) DeleteComponent(ctx context.Context, inputID string) (bool, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	componentID, _ := uuid.Parse(inputID)
	deleted, err := i.repo.Delete(ctx, componentID)
	if err != nil {
		i.log.Error("failed to delete component",
			zap.String("componentID", componentID.String()),
			zap.String("error", err.Error()))
		return false, err
	}
	return deleted, nil
}

func (i *componentService) GetComponent(ctx context.Context, inputID string) (*models.Component, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	componentID, _ := uuid.Parse(inputID)
	component, err := i.repo.Get(ctx, componentID)
	if err != nil {
		i.log.Error("failed to retrieve component",
			zap.String("componentID", componentID.String()),
			zap.String("error", err.Error()))
		return nil, err
	}

	return component, nil
}

func (i *componentService) GetAppComponents(ctx context.Context, ID string, options *models.ConnectionOptions) ([]*models.Component, *utils.Page, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	appID, _ := uuid.Parse(ID)
	components, pg, err := i.repo.ListByApp(ctx, appID, options)
	if err != nil {
		i.log.Error("failed to retrieve application's components",
			zap.String("appID", appID.String()),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}
	return components, pg, nil
}

func (i *componentService) updateComponent(ctx context.Context, input models.ComponentInput) (*models.Component, error) {
	component, err := i.GetComponent(ctx, *input.ID)
	if err != nil {
		i.log.Error("failed to retrieve component",
			zap.Any("input", input),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to retrieve component: %w", err)
	}

	// NOTE: we do not support changing region or account ID on an install
	component.Name = input.Name
	component.BuildImage = input.BuildImage
	component.Type = string(input.Type)

	if input.Config != nil {
		dbConfig, _ := component.Config.MarshalJSON()
		i.log.Info("updating component configuration",
			zap.String("existing component configuration", string(dbConfig)),
			zap.String("input component configuration", string(input.Config)))

		// convert to structs
		databaseConfig := &componentConfig.Component{}
		if err = protojson.Unmarshal([]byte(component.Config.String()), databaseConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DB JSON: %w", err)
		}
		inputConfig := &componentConfig.Component{}
		if err = protojson.Unmarshal(input.Config, inputConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input JSON: %w", err)
		}

		// calculate delta
		if inputConfig.BuildCfg != nil {
			databaseConfig.BuildCfg = inputConfig.BuildCfg
		}
		if inputConfig.DeployCfg != nil {
			databaseConfig.DeployCfg = inputConfig.DeployCfg
		}

		//convert back to byte
		var updatedConfig []byte
		updatedConfig, err = protojson.Marshal(databaseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated component configuration: %w", err)
		}
		component.Config = datatypes.JSON(updatedConfig)
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

	updatedComponent, err := i.repo.Update(ctx, component)
	if err != nil {
		i.log.Error("failed to update component",
			zap.Any("component", *component),
			zap.String("error", err.Error()))
		return nil, err
	}
	return updatedComponent, err
}

func (i *componentService) UpsertComponent(ctx context.Context, input models.ComponentInput) (*models.Component, error) {
	if input.ID != nil {
		return i.updateComponent(ctx, input)
	}

	var component models.Component
	component.Name = input.Name
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	appID, _ := uuid.Parse(input.AppID)
	component.AppID = appID
	component.BuildImage = input.BuildImage
	component.Type = string(input.Type)
	component.CreatedByID = input.CreatedByID
	if input.Config != nil {
		component.Config = datatypes.JSON(input.Config)
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

	createdComponent, err := i.repo.Create(ctx, &component)
	if err != nil {
		i.log.Error("failed to create component",
			zap.Any("component", *createdComponent),
			zap.String("error", err.Error()))
		return nil, err
	}
	return createdComponent, err
}
