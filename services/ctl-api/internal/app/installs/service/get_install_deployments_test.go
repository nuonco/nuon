package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildInstallDeploymentSortsAffectedResources(t *testing.T) {
	deploy := func(name string) app.InstallDeploy {
		return app.InstallDeploy{
			InstallComponent: app.InstallComponent{Component: app.Component{Name: name}},
		}
	}
	wf := &app.Workflow{
		ID:             "wf-1",
		Type:           app.WorkflowTypeDeployComponents,
		InstallDeploys: []app.InstallDeploy{deploy("worker"), deploy("api"), deploy("cache")},
	}
	diff := &app.InstallConfigDiff{
		Changed: []app.ComponentDiffEntry{
			{ComponentID: "comp-z", ComponentName: "zeta"},
			{ComponentID: "comp-b", ComponentName: "beta"},
		},
	}

	d := buildInstallDeployment(wf, nil, diff, map[string]app.ComponentBuild{}, nil)

	assert.Equal(t, []string{"api", "beta", "cache", "worker", "zeta"}, d.AffectedResources.Components)
}
