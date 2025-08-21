package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/hasher"
)

func (s *sync) createDockerBuildComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	obj := comp.DockerBuild

	configRequest := &models.ServiceCreateDockerBuildComponentConfigRequest{
		// DEPRECATED: BuildArgs is not used and was required for Waypoint
		AppConfigID:  s.appConfigID,
		Dependencies: comp.Dependencies,
		BuildArgs:    []string{},
		Dockerfile:   generics.ToPtr(obj.Dockerfile),
		Target:       "",
		EnvVars:      map[string]string{},
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
	}

	if obj.PublicRepo != nil {
		public := obj.PublicRepo
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    generics.ToPtr(public.Branch),
			Directory: generics.ToPtr(public.Directory),
			Repo:      generics.ToPtr(public.Repo),
		}
	}
	if obj.ConnectedRepo != nil {
		connected := obj.ConnectedRepo
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    connected.Branch,
			Directory: generics.ToPtr(connected.Directory),
			// NOTE: GitRef is not required for config sync
			Repo: generics.ToPtr(connected.Repo),
		}
	}

	configRequest.EnvVars = obj.EnvVarMap

	newChecksum, err := hasher.HashStruct(comp)
	if err != nil {
		return "", "", err
	}
	// Check if we should skip this build due to checksum match
	shouldSkip, existingConfigID, err := s.shouldSkipBuildDueToChecksum(ctx, compID, newChecksum)
	if err != nil {
		return "", "", err
	}

	if shouldSkip {
		return existingConfigID, newChecksum, nil
	}

	configRequest.Checksum = newChecksum
	cfg, err := s.apiClient.CreateDockerBuildComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, newChecksum, nil
}
