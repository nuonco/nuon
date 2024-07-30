package sync

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createHelmChartComponentConfig(ctx context.Context, resource, compID string, inp interface{}) (string, error) {
	// NOTE(jm): this logic should be updated to be handled _before_ the config gets here.
	var obj config.HelmChartComponentConfig
	if err := mapstructure.Decode(inp, &obj); err != nil {
		return "", SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("unable to parse config: %s", err.Error()),
		}
	}

	if err := obj.Parse(); err != nil {
		return "", err
	}

	configRequest := &models.ServiceCreateHelmComponentConfigRequest{
		ChartName:                generics.ToPtr(obj.ChartName),
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Values:                   map[string]string{},
		ValuesFiles:              make([]string, 0),
	}
	if obj.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    generics.ToPtr(obj.PublicRepo.Branch),
			Directory: generics.ToPtr(obj.PublicRepo.Directory),
			Repo:      generics.ToPtr(obj.PublicRepo.Repo),
		}
	} else {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    obj.ConnectedRepo.Branch,
			Directory: generics.ToPtr(obj.ConnectedRepo.Directory),
			// NOTE: GitRef is not required for config sync
			Repo: generics.ToPtr(obj.ConnectedRepo.Repo),
		}
	}
	for _, value := range obj.Values {
		configRequest.Values[value.Name] = value.Value
	}
	for k, v := range obj.ValuesMap {
		configRequest.Values[k] = v
	}

	for _, value := range obj.ValuesFiles {
		configRequest.ValuesFiles = append(configRequest.ValuesFiles, value.Contents)
	}

	cfg, err := s.apiClient.CreateHelmComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", err
	}

	return cfg.ID, nil
}
