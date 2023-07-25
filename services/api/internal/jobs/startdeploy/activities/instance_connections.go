package activities

import (
	"context"

	connectionsv1 "github.com/powertoolsdev/mono/pkg/types/components/connections/v1"
)

func (a *activities) AddConnectionsToPlan(ctx context.Context, componentID string, installID string) (*connectionsv1.Connections, error) {
	instances, err := a.instanceRepo.ListByInstall(ctx, installID)
	if err != nil {
		return nil, err
	}

	instanceConnections := []*connectionsv1.InstanceConnection{}
	for _, instance := range instances {
		if instance.ComponentID == componentID {
			// do not connect the component we are currently deploying
			// to itself
			continue
		}
		connection := &connectionsv1.InstanceConnection{
			DeployId:      instance.Deploys[0].ID,
			ComponentId:   instance.ComponentID,
			ComponentName: instance.Component.Name,
			InstallId:     instance.InstallID,
		}
		instanceConnections = append(instanceConnections, connection)
	}

	return &connectionsv1.Connections{Instances: instanceConnections}, nil
}
