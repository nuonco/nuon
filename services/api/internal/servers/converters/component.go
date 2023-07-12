package converters

import (
	"fmt"

	componentv1 "github.com/powertoolsdev/mono/pkg/types/api/component/v1"
	componentConfig "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"google.golang.org/protobuf/encoding/protojson"
)

// Component model to proto converts component domain model into component proto message
func ComponentModelToProto(component *models.Component) (*componentv1.ComponentRef, error) {
	config := &componentConfig.Component{}
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := unmarshaller.Unmarshal([]byte(component.Config.String()), config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	res := &componentv1.ComponentRef{
		Id:              component.ID,
		Name:            component.Name,
		CreatedById:     component.CreatedByID,
		ComponentConfig: config,
		UpdatedAt:       TimeToDatetime(component.UpdatedAt),
		CreatedAt:       TimeToDatetime(component.CreatedAt),
		AppId:           component.AppID,
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
