package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) createKubernetesManifestComponentConfig(
	ctx context.Context, resource, compID string, comp *config.Component,
) (string, string, error) {
	_ = comp.KubernetesManifest

	configRequest := &models.ServiceCreateKubernetesManifestComponentConfigRequest{
		AppConfigID:  s.appConfigID,
		Dependencies: comp.Dependencies,
		Checksum:     comp.Checksum,

		Namespace:     comp.KubernetesManifest.Namespace,
		Manifest:      comp.KubernetesManifest.Manifest,
		BuildTimeout:  comp.KubernetesManifest.BuildTimeout,
		DeployTimeout: comp.KubernetesManifest.DeployTimeout,
	}

	if comp.KubernetesManifest.MaxAutoRetries != nil {
		configRequest.MaxAutoRetries = int64(*comp.KubernetesManifest.MaxAutoRetries)
	}
	if comp.KubernetesManifest.Kustomize != nil {
		configRequest.Kustomize.Path = comp.KubernetesManifest.Kustomize.Path
		configRequest.Kustomize.Patches = comp.KubernetesManifest.Kustomize.Patches
		configRequest.Kustomize.EnableHelm = comp.KubernetesManifest.Kustomize.EnableHelm
		configRequest.Kustomize.LoadRestrictor = comp.KubernetesManifest.Kustomize.LoadRestrictor
	}

	// VCS configuration for kustomize sources
	if comp.KubernetesManifest.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    generics.ToPtr(comp.KubernetesManifest.PublicRepo.Branch),
			Directory: generics.ToPtr(comp.KubernetesManifest.PublicRepo.Directory),
			Repo:      generics.ToPtr(comp.KubernetesManifest.PublicRepo.Repo),
		}
	}
	if comp.KubernetesManifest.ConnectedRepo != nil {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    comp.KubernetesManifest.ConnectedRepo.Branch,
			Directory: generics.ToPtr(comp.KubernetesManifest.ConnectedRepo.Directory),
			Repo:      generics.ToPtr(comp.KubernetesManifest.ConnectedRepo.Repo),
		}
	}

	if comp.KubernetesManifest.DriftSchedule != nil {
		configRequest.DriftSchedule = *comp.KubernetesManifest.DriftSchedule
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
	}

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

	shouldSkip, existingConfigID, err := s.shouldSkipBuildDueToChecksum(ctx, compID, cmpChecksum)
	if err != nil {
		return "", "", err
	}

	if shouldSkip {
		return existingConfigID, cmpChecksum.Checksum, nil
	}

	configRequest.Checksum = cmpChecksum.Checksum
	cfg, err := s.apiClient.CreateKubernetesComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.trackBuildScheduled(compID)

	return cfg.ID, cmpChecksum.Checksum, nil
}
