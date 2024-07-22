package sync

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createTerraformModuleComponentConfig(ctx context.Context, resource, compID string, inp interface{}) (string, error) {
	var obj config.TerraformModuleComponentConfig
	if err := mapstructure.Decode(inp, &obj); err != nil {
		return "", SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("unable to parse config: %s", err.Error()),
		}
	}

	configRequest := &models.ServiceCreateTerraformModuleComponentConfigRequest{
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Variables:                map[string]string{},
		EnvVars:                  map[string]string{},
		Version:                  obj.TerraformVersion,
	}
	for _, val := range obj.Variables {
		configRequest.Variables[val.Name] = val.Value
	}
	for k, v := range obj.VarsMap {
		configRequest.Variables[k] = v
	}

	for _, val := range obj.EnvVars {
		configRequest.EnvVars[val.Name] = val.Value
	}
	for k, v := range obj.EnvVarMap {
		configRequest.EnvVars[k] = v
	}

	if obj.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    generics.ToPtr(obj.PublicRepo.Branch),
			Directory: generics.ToPtr(obj.PublicRepo.Directory),
			Repo:      generics.ToPtr(obj.PublicRepo.Repo),
		}
	} else if obj.ConnectedRepo != nil {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    obj.ConnectedRepo.Branch,
			Directory: generics.ToPtr(obj.ConnectedRepo.Directory),
			Repo:      generics.ToPtr(obj.ConnectedRepo.Repo),
		}
	}

	cfg, err := s.apiClient.CreateTerraformModuleComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", err
	}

	return cfg.ID, nil
}
