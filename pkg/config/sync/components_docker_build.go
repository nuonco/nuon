package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createDockerBuildComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, error) {
	obj := comp.DockerBuild

	configRequest := &models.ServiceCreateDockerBuildComponentConfigRequest{
		// DEPRECATED: BuildArgs is not used and was required for Waypoint
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
			// NOTE: GitRef is not required for config sync
			Repo: generics.ToPtr(connected.Repo),
		}
	}

	configRequest.EnvVars = obj.EnvVarMap

	cfg, err := s.apiClient.CreateDockerBuildComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", err
	}

	return cfg.ID, nil
}
