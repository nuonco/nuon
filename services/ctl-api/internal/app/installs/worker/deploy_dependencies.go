package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) anyDependencyInActive(ctx workflow.Context, install app.Install, installDeploy app.InstallDeploy) (string, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return "", err
	}

	l.Info("checking that all dependencies are active")
	depComponents, err := activities.AwaitGetComponentDependents(ctx, activities.GetComponentDependents{
		AppID:           install.App.ID,
		ComponentRootID: installDeploy.ComponentID,
	})
	if err != nil {
		return "", fmt.Errorf("unable to get installComponent: %w", err)
	}

	for _, dep := range depComponents {
		var depCmp *app.InstallComponent
		depCmp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   installDeploy.InstallComponent.InstallID,
			ComponentID: dep.ID,
		})

		if depCmp == nil {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("unable to get installComponent: %w", err)
		}

		if app.InstallDeployStatus(depCmp.Component.Status) != app.InstallDeployStatusActive {
			l.Error("dependdent component is not active: " + depCmp.Component.Name)
			return depCmp.ComponentID, fmt.Errorf("dependent component: %s, not active", depCmp.ID)
		}
	}

	return "", nil
}
