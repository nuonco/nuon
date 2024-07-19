package sync

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createDockerBuildComponentConfig(ctx context.Context, resource, compID string, inp interface{}) (string, error) {
	var obj config.DockerBuildComponentConfig
	if err := mapstructure.Decode(inp, &obj); err != nil {
		return "", SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("unable to parse config: %s", err.Error()),
		}
	}

	configRequest := &models.ServiceCreateDockerBuildComponentConfigRequest{
		BuildArgs:  []string{},
		Dockerfile: generics.ToPtr(obj.Dockerfile),
		Target:     "",
		EnvVars:    map[string]string{},
	}

	if obj.PublicRepo != nil {
		public := obj.PublicRepo
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    generics.ToPtr(public.Branch),
			Directory: generics.ToPtr(public.Directory),
			Repo:      generics.ToPtr(public.Repo),
		}
	} else {
		connected := obj.ConnectedRepo
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    connected.Branch,
			Directory: generics.ToPtr(connected.Directory),
			Repo:      generics.ToPtr(connected.Repo),
		}
	}
	for _, envVar := range obj.EnvVars {
		configRequest.EnvVars[envVar.Name] = envVar.Value
	}

	cfg, err := s.apiClient.CreateDockerBuildComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", err
	}

	return cfg.ID, nil
}
