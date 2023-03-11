package converters

import (
	"github.com/powertoolsdev/mono/services/api/internal/models"
	installv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/install/v1"
)

// Install model to proto converts install domain model into install proto message
func InstallModelToProto(install *models.Install) *installv1.Install {
	return &installv1.Install{
		Id:   install.ID.String(),
		Name: install.Name,
		Settings: &installv1.Install_AwsSettings{
			AwsSettings: &installv1.AwsSettings{
				Region: AwsRegionToProto(install.AWSSettings.Region),
				Role:   install.AWSSettings.IamRoleArn,
			},
		},
		CreatedById: install.CreatedByID,
		CreatedAt:   TimeToDatetime(install.CreatedAt),
		UpdatedAt:   TimeToDatetime(install.UpdatedAt),
	}
}

// InstallModelsToProtos converts a slice of install models to protos
func InstallModelsToProtos(installs []*models.Install) []*installv1.Install {
	protos := make([]*installv1.Install, len(installs))
	for idx, install := range installs {
		protos[idx] = InstallModelToProto(install)
	}

	return protos
}
