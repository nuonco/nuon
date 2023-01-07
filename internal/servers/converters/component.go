package converters

import (
	"github.com/powertoolsdev/api/internal/models"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

// Component model to proto converts component domain model into component proto message
func ComponentModelToProto(component *models.Component) *componentv1.ComponentRef {
	return &componentv1.ComponentRef{
		Id:          component.ID.String(),
		Name:        component.Name,
		CreatedById: component.CreatedByID,
		UpdatedAt:   TimeToDatetime(component.UpdatedAt),
		CreatedAt:   TimeToDatetime(component.CreatedAt),
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
