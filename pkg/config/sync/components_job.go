package sync

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createJobComponentConfig(ctx context.Context, resource, compID string, obj interface{}) (string, error) {
	var containerImage config.JobComponentConfig
	if err := mapstructure.Decode(obj, &containerImage); err != nil {
		return "", SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("unable to parse config: %s", err.Error()),
		}
	}

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
