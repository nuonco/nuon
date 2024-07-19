package sync

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createContainerImageComponentConfig(ctx context.Context, resource, compID string, obj interface{}) (string, error) {
	var containerImage config.ExternalImageComponentConfig
	if err := mapstructure.Decode(obj, &containerImage); err != nil {
		return "", SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("unable to parse config: %s", err.Error()),
		}
	}

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
