package converters

import (
	deploymentv1 "github.com/powertoolsdev/mono/pkg/types/api/deployment/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

// Deployment model to proto converts deployment domain model into deployment proto message
func DeploymentModelToProto(deployment *models.Deployment) *deploymentv1.Deployment {
	return &deploymentv1.Deployment{
		Id:           deployment.ID.String(),
		CommitAuthor: deployment.CommitAuthor,
		CommitHash:   deployment.CommitHash,
		ComponentId:  deployment.ComponentID.String(),
		CreatedById:  deployment.CreatedByID,
		// TODO: return []string of InstallIDs
		UpdatedAt: TimeToDatetime(deployment.UpdatedAt),
		CreatedAt: TimeToDatetime(deployment.CreatedAt),
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
