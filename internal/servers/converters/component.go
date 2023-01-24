package converters

import (
	"github.com/powertoolsdev/api/internal/models"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

func ComponentTypeToProto(input models.ComponentType) componentv1.ComponentType {
	switch input {
	case models.ComponentTypeGithubRepo:
		return componentv1.ComponentType_COMPONENT_TYPE_GITHUB_REPO
	case models.ComponentTypePublicImage:
		return componentv1.ComponentType_COMPONENT_TYPE_PUBLIC_IMAGE
	case models.ComponentTypeHelm:
		return componentv1.ComponentType_COMPONENT_TYPE_HELM
	case models.ComponentTypeTerraform:
		return componentv1.ComponentType_COMPONENT_TYPE_TERRAFORM
	}

	return componentv1.ComponentType_COMPONENT_TYPE_UNSPECIFIED
}

// Component model to proto converts component domain model into component proto message
func ComponentModelToProto(component *models.Component) *componentv1.ComponentRef {
	return &componentv1.ComponentRef{
		Id:          component.ID.String(),
		Name:        component.Name,
		CreatedById: component.CreatedByID,
		BuildImage:  component.BuildImage,
		Type:        ComponentTypeToProto(models.ComponentType(component.Type)),
		VcsConfig: &componentv1.ComponentRef_GithubConfig{
			GithubConfig: &componentv1.GithubConfig{
				Branch:    component.GithubConfig.Branch,
				Directory: component.GithubConfig.Directory,
				Repo:      component.GithubConfig.Repo,
				RepoOwner: component.GithubConfig.RepoOwner,
			},
		},
		UpdatedAt: TimeToDatetime(component.UpdatedAt),
		CreatedAt: TimeToDatetime(component.CreatedAt),
	}
}

// ComponentModelsToProtos converts a slice of component models to protos
func ComponentModelsToProtos(components []*models.Component) []*componentv1.ComponentRef {
	protos := make([]*componentv1.ComponentRef, len(components))
	for idx, component := range components {
		protos[idx] = ComponentModelToProto(component)
	}

	return protos
}
