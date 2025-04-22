package sync

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createJobComponentConfig(ctx context.Context, resource, compID string, obj *config.Component) (string, string, error) {
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

	// NOTE: we don't want to make a checksum with the app config id since that can change
	configRequest.AppConfigID = s.appConfigID

	cfg, err := s.apiClient.CreateJobComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, requestChecksum, nil
}
