package sync

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

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

func (s *sync) syncComponentConfig(ctx context.Context, comp *config.Component, resource, compID string) (string, error) {
	// TODO(jm): this method can now use the Parse method to get an actual component object, simplifying the map
	// decoding everywhere in this package.

	methods := map[models.AppComponentType]func(context.Context, string, string, *config.Component) (string, error){
		models.AppComponentTypeHelmChart:       s.createHelmChartComponentConfig,
		models.AppComponentTypeTerraformModule: s.createTerraformModuleComponentConfig,
		models.AppComponentTypeDockerBuild:     s.createDockerBuildComponentConfig,
		models.AppComponentTypeExternalImage:   s.createContainerImageComponentConfig,
		models.AppComponentTypeJob:             s.createJobComponentConfig,
	}
	method, ok := methods[comp.Type.APIType()]
	if !ok {
		return "", SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("invalid type %s", comp.Type),
		}
	}

	cfgID, err := method(ctx, resource, compID, comp)
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

func (s *sync) syncComponent(ctx context.Context, resource string, comp *config.Component) (string, error) {
	var isNew bool
	apiComp, err := s.getComponent(ctx, comp.Name, comp.Type.APIType())
	if err != nil {
		if !nuon.IsNotFound(err) {
			return "", err
		}

		isNew = true
		apiComp, err = s.apiClient.CreateComponent(ctx, s.appID, &models.ServiceCreateComponentRequest{
			Dependencies: comp.Dependencies,
			Name:         generics.ToPtr(comp.Name),
			VarName:      comp.VarName,
		})
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	if !isNew {
		_, err = s.apiClient.UpdateComponent(ctx, apiComp.ID, &models.ServiceUpdateComponentRequest{
			Dependencies: comp.Dependencies,
			VarName:      comp.VarName,
			Name:         generics.ToPtr(comp.Name),
		})
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	cfgID, err := s.syncComponentConfig(ctx, comp, resource, apiComp.ID)
	if err != nil {
		if isNew {
			s.cleanupComponent(ctx, apiComp.ID)
		}

		return "", err
	}

	s.state.ComponentIDs = append(s.state.ComponentIDs, componentState{
		Name:     apiComp.ID,
		Type:     comp.Type.APIType(),
		ID:       apiComp.ID,
		ConfigID: cfgID,
	})
	return apiComp.ID, nil
}
