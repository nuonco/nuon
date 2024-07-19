package sync

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) parseMinComponent(cfg interface{}) (config.MinComponent, error) {
	var minComponent config.MinComponent
	if err := mapstructure.Decode(cfg, &minComponent); err != nil {
		return config.MinComponent{}, err
	}

	return minComponent, nil
}

func (s *sync) getComponent(ctx context.Context, name string, typ models.AppComponentType) (*models.AppComponent, error) {
	comp, err := s.apiClient.GetAppComponent(ctx, s.appID, name)
	if err != nil {
		return nil, err
	}

	if typ != comp.Type {
		return nil, SyncErr{
			Resource:    fmt.Sprintf("%s component", typ),
			Description: "previous component was found with a different type",
		}
	}

	return comp, nil
}

func (s *sync) syncComponentConfig(ctx context.Context, minComp config.MinComponent, cfg interface{}, resource, compID string) (string, error) {
	// create config version
	methods := map[models.AppComponentType]func(context.Context, string, string, interface{}) (string, error){
		models.AppComponentTypeHelmChart:       s.createHelmChartComponentConfig,
		models.AppComponentTypeTerraformModule: s.createTerraformModuleComponentConfig,
		models.AppComponentTypeDockerBuild:     s.createDockerBuildComponentConfig,
		models.AppComponentTypeExternalImage:   s.createContainerImageComponentConfig,
		models.AppComponentTypeJob:             s.createJobComponentConfig,
	}
	method, ok := methods[minComp.APIType()]
	if !ok {
		return "", SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("invalid type %s", minComp.Type),
		}
	}

	cfgID, err := method(ctx, resource, compID, cfg)
	if err != nil {
		return "", SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return cfgID, nil
}

func (s *sync) cleanupComponent(ctx context.Context, compID string) {
	_, err := s.apiClient.DeleteComponent(ctx, compID)
	if err != nil {
		fmt.Println("unable to delete component after config: %w", err)
	}
}

func (s *sync) syncComponent(ctx context.Context, resource string, cfg interface{}, dependencies []string) (string, error) {
	minComponent, err := s.parseMinComponent(cfg)
	if err != nil {
		return "", err
	}

	var isNew bool
	comp, err := s.getComponent(ctx, minComponent.Name, minComponent.APIType())
	if err != nil {
		if !nuon.IsNotFound(err) {
			return "", err
		}

		isNew = true
		comp, err = s.apiClient.CreateComponent(ctx, s.appID, &models.ServiceCreateComponentRequest{
			Dependencies: dependencies,
			Name:         generics.ToPtr(minComponent.Name),
			VarName:      minComponent.VarName,
		})
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	if !isNew {
		_, err = s.apiClient.UpdateComponent(ctx, comp.ID, &models.ServiceUpdateComponentRequest{
			Dependencies: dependencies,
			VarName:      minComponent.VarName,
			Name:         generics.ToPtr(minComponent.Name),
		})
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	cfgID, err := s.syncComponentConfig(ctx, minComponent, cfg, resource, comp.ID)
	if err != nil {
		if isNew {
			s.cleanupComponent(ctx, comp.ID)
		}

		return "", err
	}

	s.state.ComponentIDs = append(s.state.ComponentIDs, componentState{
		Name:     comp.ID,
		Type:     minComponent.APIType(),
		ID:       comp.ID,
		ConfigID: cfgID,
	})
	return comp.ID, nil
}
