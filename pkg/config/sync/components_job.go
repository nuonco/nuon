package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createJobComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	containerImage := comp.Job

	envVars := make(map[string]string, 0)
	for _, value := range containerImage.EnvVars {
		envVars[value.Name] = value.Value
	}
	for k, v := range containerImage.EnvVarMap {
		envVars[k] = v
	}

	configRequest := &models.ServiceCreateJobComponentConfigRequest{
		AppConfigID: s.appConfigID,
		Args:        containerImage.Args,
		Cmd:         containerImage.Cmd,
		EnvVars:     envVars,
		ImageURL:    generics.ToPtr(containerImage.ImageURL),
		Tag:         generics.ToPtr(containerImage.Tag),
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
	cfg, err := s.apiClient.CreateJobComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, cmpChecksum.Checksum, nil
}
