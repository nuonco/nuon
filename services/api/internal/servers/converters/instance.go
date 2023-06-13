package converters

import (
	instancev1 "github.com/powertoolsdev/mono/pkg/types/api/instance/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

// Instance model to proto converts instance domain model into instance proto message
func InstanceModelToProto(instance *models.Instance) *instancev1.Instance {
	return &instancev1.Instance{
		Id:          instance.ID,
		BuildId:     instance.BuildID,
		ComponentId: instance.ComponentID,
	}
}

// InstanceModelsToProtos converts a slice of instance models to protos
func InstanceModelsToProtos(instances []*models.Instance) []*instancev1.Instance {
	protos := make([]*instancev1.Instance, len(instances))
	for idx, instance := range instances {
		protos[idx] = InstanceModelToProto(instance)
	}

	return protos
}
