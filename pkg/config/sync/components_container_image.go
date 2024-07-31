package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createContainerImageComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, error) {
	containerImage := comp.ExternalImage

	configRequest := &models.ServiceCreateExternalImageComponentConfigRequest{}
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

	cfg, err := s.apiClient.CreateExternalImageComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", err
	}

	return cfg.ID, nil
}
