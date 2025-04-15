package sync

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createDockerBuildComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	obj := comp.DockerBuild

	configRequest := &models.ServiceCreateDockerBuildComponentConfigRequest{
		AppConfigID: s.appConfigID,
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

	requestChecksum, err := s.getChecksum(configRequest)
	if err != nil {
		return "", "", err
	}

	cmpBuild, err := s.apiClient.GetComponentLatestBuild(ctx, compID)
	if err != nil && !nuon.IsNotFound(err) {
		return "", "", err
	}

	doChecksumCompare := true
	if cmpBuild != nil && cmpBuild.Status == "error" {
		doChecksumCompare = false
	}

	if doChecksumCompare {
		prevComponentState := s.getComponentStateById(compID)
		if prevComponentState != nil && prevComponentState.Checksum == requestChecksum {
			return prevComponentState.ConfigID, requestChecksum, nil
		}
	}

	cfg, err := s.apiClient.CreateDockerBuildComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, requestChecksum, nil
}
