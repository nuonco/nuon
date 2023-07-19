package activities

import (
	"context"

	connectionsv1 "github.com/powertoolsdev/mono/pkg/types/components/connections/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func (a *activities) AddConnectionsToPlan(ctx context.Context, buildPlan *planv1.Plan) (*planv1.Plan, error) {
	waypointPlan := buildPlan.GetWaypointPlan()

	instances, err := a.instanceRepo.ListByInstall(ctx, waypointPlan.Metadata.InstallId)
	if err != nil {
		return nil, err
	}

	instanceConnections := []*connectionsv1.InstanceConnection{}
	for _, instance := range instances {
		connection := &connectionsv1.InstanceConnection{
			DeployId:      waypointPlan.Metadata.DeploymentId,
			ComponentId:   instance.ComponentID,
			ComponentName: instance.Component.Name,
			InstallId:     instance.InstallID,
		}
		instanceConnections = append(instanceConnections, connection)
	}

	waypointPlan.Component.Connections = &connectionsv1.Connections{
		Instances: instanceConnections,
	}

	buildPlan.Actual = &planv1.Plan_WaypointPlan{
		WaypointPlan: waypointPlan,
	}

	return buildPlan, nil
}
