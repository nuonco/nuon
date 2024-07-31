package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createJobComponentConfig(ctx context.Context, resource, compID string, obj *config.Component) (string, error) {
	containerImage := obj.Job

	envVars := make(map[string]string, 0)
	for _, value := range containerImage.EnvVars {
		envVars[value.Name] = value.Value
	}
	for k, v := range containerImage.EnvVarMap {
		envVars[k] = v
	}

	configRequest := &models.ServiceCreateJobComponentConfigRequest{
		Args:     containerImage.Args,
		Cmd:      containerImage.Cmd,
		EnvVars:  envVars,
		ImageURL: generics.ToPtr(containerImage.ImageURL),
		Tag:      generics.ToPtr(containerImage.Tag),
	}

	cfg, err := s.apiClient.CreateJobComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", err
	}

	return cfg.ID, nil
}
