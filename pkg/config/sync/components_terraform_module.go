package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createTerraformModuleComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	obj := comp.TerraformModule

	configRequest := &models.ServiceCreateTerraformModuleComponentConfigRequest{
		AppConfigID:              s.appConfigID,
		Dependencies:             comp.Dependencies,
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Variables:                map[string]string{},
		EnvVars:                  map[string]string{},
		VariablesFiles:           make([]string, 0),
		Version:                  obj.TerraformVersion,
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
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
	for _, v := range obj.VariablesFiles {
		configRequest.VariablesFiles = append(configRequest.VariablesFiles, v.Contents)
	}

	if obj.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    generics.ToPtr(obj.PublicRepo.Branch),
			Directory: generics.ToPtr(obj.PublicRepo.Directory),
			Repo:      generics.ToPtr(obj.PublicRepo.Repo),
		}
	}
	if obj.ConnectedRepo != nil {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    obj.ConnectedRepo.Branch,
			Directory: generics.ToPtr(obj.ConnectedRepo.Directory),
			// NOTE: GitRef is not required for config sync
			Repo: generics.ToPtr(obj.ConnectedRepo.Repo),
		}
	}

	cmpChecksum, err := s.generateComponentChecksun(ctx, comp)
	if err != nil {
		return "", "", err
	}
	shouldSkip, existingConfigID, err := s.shouldSkipBuildDueToChecksum(ctx, compID, cmpChecksum)
	if err != nil {
		return "", "", err
	}

	if shouldSkip {
		return existingConfigID, cmpChecksum.Checksum, nil
	}

	configRequest.Checksum = cmpChecksum.Checksum
	cfg, err := s.apiClient.CreateTerraformModuleComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, cmpChecksum.Checksum, nil
}
