package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func (s *syncer) createDockerBuildComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	obj := comp.DockerBuild

	configRequest := &models.ServiceCreateDockerBuildComponentConfigRequest{
		// DEPRECATED: BuildArgs is not used and was required for Waypoint
		AppConfigID:   s.appConfigID,
		Dependencies:  comp.Dependencies,
		BuildArgs:     []string{},
		Dockerfile:    generics.ToPtr(obj.Dockerfile),
		Target:        "",
		EnvVars:       map[string]string{},
		BuildTimeout:  obj.BuildTimeout,
		DeployTimeout: obj.DeployTimeout,
	}

	if obj.MaxAutoRetries != nil {
		configRequest.MaxAutoRetries = int64(*obj.MaxAutoRetries)
	}
	if obj.SkipNoops != nil {
		configRequest.SkipNoops = *obj.SkipNoops
	}
	if comp.Toggleable != nil {
		configRequest.Toggleable = *comp.Toggleable
	}
	if comp.DefaultEnabled != nil {
		configRequest.DefaultEnabled = *comp.DefaultEnabled
	}
	if obj.AutoApproveOnPoliciesPassing != nil {
		configRequest.AutoApproveOnPoliciesPassing = *obj.AutoApproveOnPoliciesPassing
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

	if len(comp.OperationRoles) > 0 {
		configRequest.OperationRoles = make(map[string]string)
		for _, opRole := range comp.OperationRoles {
			configRequest.OperationRoles[string(opRole.Operation)] = opRole.RoleName
		}
	}

	cmpChecksum, err := s.generateComponentChecksun(ctx, comp)
	if err != nil {
		return "", "", err
	}
	// Check if we should skip this build due to checksum match
	shouldSkip, existingConfigID, err := s.shouldSkipBuildDueToChecksum(ctx, compID, cmpChecksum)
	if err != nil {
		return "", "", err
	}

	if shouldSkip {
		return existingConfigID, cmpChecksum.Checksum, nil
	}

	configRequest.Checksum = cmpChecksum.Checksum
	cfg, err := s.apiClient.CreateDockerBuildComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.trackBuildScheduled(compID)

	return cfg.ID, cmpChecksum.Checksum, nil
}
