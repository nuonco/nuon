package apisyncer

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type SyncComponentConfigResponse struct {
	CfgID, RequestChecksum string
}

func (s *syncer) getComponent(ctx context.Context, name string, typ models.AppComponentType) (*models.AppComponent, error) {
	comp, err := s.apiClient.GetAppComponent(ctx, s.appID, name)
	if err != nil {
		return nil, err
	}

	if typ != comp.Type && !generics.SliceContains(comp.Type, []models.AppComponentType{
		models.AppComponentTypeUnknown,
		"",
	}) {
		return nil, sync.SyncErr{
			Resource:    fmt.Sprintf("%s component", typ),
			Description: "previous component was found with a different type",
		}
	}

	return comp, nil
}

func (s *syncer) syncComponentConfig(ctx context.Context, comp *config.Component, resource, compID string) (SyncComponentConfigResponse, error) {
	// TODO(jm): this method can now use the Parse method to get an actual component object, simplifying the map
	// decoding everywhere in this package.

	methods := map[models.AppComponentType]func(context.Context, string, string, *config.Component) (string, string, error){
		models.AppComponentTypeHelmChart:          s.createHelmChartComponentConfig,
		models.AppComponentTypeTerraformModule:    s.createTerraformModuleComponentConfig,
		models.AppComponentTypeDockerBuild:        s.createDockerBuildComponentConfig,
		models.AppComponentTypeExternalImage:      s.createContainerImageComponentConfig,
		models.AppComponentTypeJob:                s.createJobComponentConfig,
		models.AppComponentTypeKubernetesManifest: s.createKubernetesManifestComponentConfig,
		models.AppComponentTypePulumi:             s.createPulumiComponentConfig,
	}
	method, ok := methods[comp.Type.APIType()]
	if !ok {
		return SyncComponentConfigResponse{
				CfgID:           "",
				RequestChecksum: "",
			}, sync.SyncErr{
				Resource:    resource,
				Description: fmt.Sprintf("invalid type %s", comp.Type),
			}
	}

	cfgID, requestChecksum, err := method(ctx, resource, compID, comp)
	if err != nil {
		return SyncComponentConfigResponse{
				CfgID:           "",
				RequestChecksum: "",
			}, sync.SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
	}

	return SyncComponentConfigResponse{
		CfgID:           cfgID,
		RequestChecksum: requestChecksum,
	}, nil
}

func (s *syncer) cleanupComponent(ctx context.Context, compID string) error {
	_, err := s.apiClient.DeleteComponent(ctx, compID)
	if err != nil {
		return fmt.Errorf("unable to delete component after config: %w", err)
	}

	return nil
}

func (s *syncer) ensureComponent(ctx context.Context, resource string, comp *config.Component) error {
	_, err := s.getComponent(ctx, comp.Name, comp.Type.APIType())
	if err != nil {
		if !nuon.IsNotFound(err) {
			return err
		}
	}

	if err == nil {
		return nil
	}

	_, err = s.apiClient.CreateComponent(ctx, s.appID, &models.ServiceCreateComponentRequest{
		Name:   generics.ToPtr(comp.Name),
		Labels: comp.Labels,
	})
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}

func (s *syncer) syncComponent(ctx context.Context, resource string, comp *config.Component) (string, error) {
	apiComp, err := s.getComponent(ctx, comp.Name, comp.Type.APIType())
	if err != nil {
		return "", err
	}

	_, err = s.apiClient.UpdateComponent(ctx, apiComp.ID, &models.ServiceUpdateComponentRequest{
		Dependencies: comp.Dependencies,
		VarName:      comp.VarName,
		Name:         generics.ToPtr(comp.Name),
		Labels:       comp.Labels,
	})
	if err != nil {
		return "", sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	resp, err := s.syncComponentConfig(ctx, comp, resource, apiComp.ID)
	if err != nil {
		return "", err
	}

	s.appendComponentState(sync.ComponentState{
		Name:     apiComp.Name,
		Type:     comp.Type.APIType(),
		ID:       apiComp.ID,
		ConfigID: resp.CfgID,
		Checksum: resp.RequestChecksum,
	})

	return apiComp.ID, nil
}
