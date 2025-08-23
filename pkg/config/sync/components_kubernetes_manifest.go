package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/config"
)

func (s *sync) createKubernetesManifestComponentConfig(
	ctx context.Context, resource, compID string, comp *config.Component,
) (string, string, error) {
	_ = comp.KubernetesManifest

	configRequest := &models.ServiceCreateKubernetesManifestComponentConfigRequest{
		AppConfigID:  s.appConfigID,
		Dependencies: comp.Dependencies,
		Checksum:     comp.Checksum,

		Namespace: comp.KubernetesManifest.Namespace,
		Manifest:  comp.KubernetesManifest.Manifest,
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
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

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, cmpChecksum.Checksum, nil
}
