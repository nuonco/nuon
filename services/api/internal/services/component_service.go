package services

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	componentConfig "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/utils"
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
	log     *zap.Logger
	repo    repos.ComponentRepo
	appRepo repos.AppRepo
}

var _ ComponentService = (*componentService)(nil)

func NewComponentService(db *gorm.DB, log *zap.Logger) ComponentService {
	componentRepo := repos.NewComponentRepo(db)
	return &componentService{
		log:     log,
		repo:    componentRepo,
		appRepo: repos.NewAppRepo(db),
	}
}

func (i *componentService) DeleteComponent(ctx context.Context, componentID string) (bool, error) {
	deleted, err := i.repo.Delete(ctx, componentID)
	if err != nil {
		i.log.Error("failed to delete component",
			zap.String("componentID", componentID),
			zap.String("error", err.Error()))
		return false, err
	}
	return deleted, nil
}

func (i *componentService) GetComponent(ctx context.Context, componentID string) (*models.Component, error) {
	component, err := i.repo.Get(ctx, componentID)
	if err != nil {
		i.log.Error("failed to retrieve component",
			zap.String("componentID", componentID),
			zap.String("error", err.Error()))
		return nil, err
	}

	return component, nil
}

func (i *componentService) GetAppComponents(ctx context.Context, appID string, options *models.ConnectionOptions) ([]*models.Component, *utils.Page, error) {
	components, pg, err := i.repo.ListByApp(ctx, appID, options)
	if err != nil {
		i.log.Error("failed to retrieve application's components",
			zap.String("appID", appID),
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

	// check if app exists
	_, err := i.appRepo.Get(ctx, input.AppID)
	if err != nil {
		i.log.Error("failed to get app",
			zap.String("appID", input.AppID),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	var component models.Component
	component.ID = domains.NewComponentID()
	component.Name = input.Name
	component.AppID = input.AppID
	component.CreatedByID = input.CreatedByID
	if input.Config != nil {
		component.Config = datatypes.JSON(input.Config)
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
