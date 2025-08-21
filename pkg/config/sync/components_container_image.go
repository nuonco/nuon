package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/hasher"
)

func (s *sync) createContainerImageComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	containerImage := comp.ExternalImage

	configRequest := &models.ServiceCreateExternalImageComponentConfigRequest{
		AppConfigID:  s.appConfigID,
		Dependencies: comp.Dependencies,
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
	}

	if containerImage.AWSECRImageConfig != nil {
		configRequest.ImageURL = generics.ToPtr(containerImage.AWSECRImageConfig.ImageURL)
		configRequest.Tag = generics.ToPtr(containerImage.AWSECRImageConfig.Tag)
		configRequest.AwsEcrImageConfig = &models.ServiceAwsECRImageConfigRequest{
			AwsRegion:  containerImage.AWSECRImageConfig.AWSRegion,
			IamRoleArn: containerImage.AWSECRImageConfig.IAMRoleARN,
		}
	} else if containerImage.PublicImageConfig != nil {
		configRequest.ImageURL = generics.ToPtr(containerImage.PublicImageConfig.ImageURL)
		configRequest.Tag = generics.ToPtr(containerImage.PublicImageConfig.Tag)
	}

	newChecksum, err := hasher.HashStruct(comp)
	if err != nil {
		return "", "", err
	}
	// Check if we should skip this build due to checksum match
	shouldSkip, existingConfigID, err := s.shouldSkipBuildDueToChecksum(ctx, compID, newChecksum)
	if err != nil {
		return "", "", err
	}

	if shouldSkip {
		return existingConfigID, newChecksum, nil
	}

	configRequest.Checksum = newChecksum
	cfg, err := s.apiClient.CreateExternalImageComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, newChecksum, nil
}
