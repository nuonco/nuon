package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func (s *syncer) createPulumiComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	obj := comp.Pulumi

	configRequest := &models.ServiceCreatePulumiComponentConfigRequest{
		AppConfigID:              s.appConfigID,
		Dependencies:             comp.Dependencies,
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Config:                   map[string]string{},
		EnvVars:                  map[string]string{},
		Runtime:                  generics.ToPtr(obj.Runtime),
		Version:                  obj.PulumiVersion,
		BuildTimeout:             obj.BuildTimeout,
		DeployTimeout:            obj.DeployTimeout,
	}

	if obj.MaxAutoRetries != nil {
		configRequest.MaxAutoRetries = int64(*obj.MaxAutoRetries)
	}
	if obj.SkipNoops != nil {
		configRequest.SkipNoops = *obj.SkipNoops
	}
	if obj.AutoApproveOnPoliciesPassing != nil {
		configRequest.AutoApproveOnPoliciesPassing = *obj.AutoApproveOnPoliciesPassing
	}
	if obj.DriftSchedule != nil {
		configRequest.DriftSchedule = *obj.DriftSchedule
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
	}

	for k, v := range obj.ConfigMap {
		configRequest.Config[k] = v
	}

	for k, v := range obj.EnvVarMap {
		configRequest.EnvVars[k] = v
	}

	if len(comp.OperationRoles) > 0 {
		configRequest.OperationRoles = make(map[string]string)
		for _, opRole := range comp.OperationRoles {
			configRequest.OperationRoles[string(opRole.Operation)] = opRole.RoleName
		}
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
			Repo:      generics.ToPtr(obj.ConnectedRepo.Repo),
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
	cfg, err := s.apiClient.CreatePulumiComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, cmpChecksum.Checksum, nil
}
