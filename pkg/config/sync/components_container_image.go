package sync

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createContainerImageComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	containerImage := comp.ExternalImage

	configRequest := &models.ServiceCreateExternalImageComponentConfigRequest{
		Dependencies: comp.Dependencies,
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

	cfg, err := s.apiClient.CreateExternalImageComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, requestChecksum, nil
}
