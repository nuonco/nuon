package protos

import (
	connectionsv1 "github.com/powertoolsdev/mono/pkg/types/components/connections/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) toConnections(installDeploys []app.InstallDeploy) *connectionsv1.Connections {
	instances := make([]*connectionsv1.InstanceConnection, 0)
	for _, deploy := range installDeploys {
		instances = append(instances, &connectionsv1.InstanceConnection{
			ComponentId:   deploy.InstallComponent.ComponentID,
			InstallId:     deploy.InstallComponent.InstallID,
			DeployId:      deploy.ID,
			ComponentName: deploy.InstallComponent.Component.Name,
			VarName:       deploy.InstallComponent.Component.ResolvedVarName,
		})
	}

	return &connectionsv1.Connections{
		Instances: instances,
	}
}
