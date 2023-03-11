package converters

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	componentv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/component/v1"
	componentConfig "github.com/powertoolsdev/mono/pkg/protos/components/generated/types/component/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Component model to proto converts component domain model into component proto message
func ComponentModelToProto(component *models.Component) (*componentv1.ComponentRef, error) {
	config := &componentConfig.Component{}
	if err := protojson.Unmarshal([]byte(component.Config.String()), config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	res := &componentv1.ComponentRef{
		Id:              component.ID.String(),
		Name:            component.Name,
		CreatedById:     component.CreatedByID,
		ComponentConfig: config,
		UpdatedAt:       TimeToDatetime(component.UpdatedAt),
		CreatedAt:       TimeToDatetime(component.CreatedAt),
	}

	return res, nil
}

// ComponentModelsToProtos converts a slice of component models to protos
func ComponentModelsToProtos(components []*models.Component) ([]*componentv1.ComponentRef, error) {
	protos := make([]*componentv1.ComponentRef, len(components))
	for idx, component := range components {
		protoComponent, err := ComponentModelToProto(component)
		if err != nil {
			return nil, err
		}
		protos[idx] = protoComponent
	}

	return protos, nil
}
