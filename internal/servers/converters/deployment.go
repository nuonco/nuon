package converters

import (
	"github.com/powertoolsdev/api/internal/models"
	deploymentv1 "github.com/powertoolsdev/protos/api/generated/types/deployment/v1"
)

// Deployment model to proto converts deployment domain model into deployment proto message
func DeploymentModelToProto(deployment *models.Deployment) *deploymentv1.Deployment {
	return &deploymentv1.Deployment{
		Id:          deployment.ID.String(),
		CreatedById: deployment.CreatedByID,
		UpdatedAt:   TimeToDatetime(deployment.UpdatedAt),
		CreatedAt:   TimeToDatetime(deployment.CreatedAt),
	}
}

// DeploymentModelsToProtos converts a slice of deployment models to protos
func DeploymentModelsToProtos(deployments []*models.Deployment) []*deploymentv1.Deployment {
	protos := make([]*deploymentv1.Deployment, len(deployments))
	for idx, deployment := range deployments {
		protos[idx] = DeploymentModelToProto(deployment)
	}

	return protos
}
