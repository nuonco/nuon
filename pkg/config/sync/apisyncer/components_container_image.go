package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func (s *syncer) createContainerImageComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	containerImage := comp.ExternalImage

	configRequest := &models.ServiceCreateExternalImageComponentConfigRequest{
		AppConfigID:   s.appConfigID,
		Dependencies:  comp.Dependencies,
		BuildTimeout:  containerImage.BuildTimeout,
		DeployTimeout: containerImage.DeployTimeout,
	}

	if comp.Toggleable != nil {
		configRequest.Toggleable = *comp.Toggleable
	}
	if comp.DefaultEnabled != nil {
		configRequest.DefaultEnabled = *comp.DefaultEnabled
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
	}

	if containerImage.AWSECRImageConfig != nil {
		configRequest.ImageURL = generics.ToPtr(containerImage.AWSECRImageConfig.ImageURL)
		configRequest.Tag = containerImage.AWSECRImageConfig.Tag
		configRequest.UpdatePolicy = containerImage.AWSECRImageConfig.UpdatePolicy
		configRequest.AwsEcrImageConfig = &models.ServiceAwsECRImageConfigRequest{
			AwsRegion:  containerImage.AWSECRImageConfig.AWSRegion,
			IamRoleArn: containerImage.AWSECRImageConfig.IAMRoleARN,
		}
	} else if containerImage.GCPGARImageConfig != nil {
		configRequest.ImageURL = generics.ToPtr(containerImage.GCPGARImageConfig.ImageURL)
		configRequest.Tag = containerImage.GCPGARImageConfig.Tag
		configRequest.UpdatePolicy = containerImage.GCPGARImageConfig.UpdatePolicy
		configRequest.GcpGarImageConfig = &models.ServiceGcpGARImageConfigRequest{
			GcpProjectID:             containerImage.GCPGARImageConfig.GCPProjectID,
			GcpRegion:                containerImage.GCPGARImageConfig.GCPRegion,
			ServiceAccountEmail:      containerImage.GCPGARImageConfig.ServiceAccountEmail,
			WorkloadIdentityProvider: containerImage.GCPGARImageConfig.WorkloadIdentityProvider,
		}
	} else if containerImage.AzureACRImageConfig != nil {
		configRequest.ImageURL = generics.ToPtr(containerImage.AzureACRImageConfig.ImageURL)
		configRequest.Tag = containerImage.AzureACRImageConfig.Tag
		configRequest.UpdatePolicy = containerImage.AzureACRImageConfig.UpdatePolicy
		configRequest.AzureAcrImageConfig = &models.ServiceAzureACRImageConfigRequest{
			RegistryURL: containerImage.AzureACRImageConfig.RegistryURL,
			TenantID:    containerImage.AzureACRImageConfig.TenantID,
			ClientID:    containerImage.AzureACRImageConfig.ClientID,
		}
	} else if containerImage.PublicImageConfig != nil {
		configRequest.ImageURL = generics.ToPtr(containerImage.PublicImageConfig.ImageURL)
		configRequest.Tag = containerImage.PublicImageConfig.Tag
		configRequest.UpdatePolicy = containerImage.PublicImageConfig.UpdatePolicy
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
	// Check if we should skip this build due to checksum match
	shouldSkip, existingConfigID, err := s.shouldSkipBuildDueToChecksum(ctx, compID, cmpChecksum)
	if err != nil {
		return "", "", err
	}

	if shouldSkip {
		return existingConfigID, cmpChecksum.Checksum, nil
	}

	configRequest.Checksum = cmpChecksum.Checksum
	cfg, err := s.apiClient.CreateExternalImageComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.trackBuildScheduled(compID)

	return cfg.ID, cmpChecksum.Checksum, nil
}
