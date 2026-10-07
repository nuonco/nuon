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

func TestApplyPinnedConfigBranchFillsProvisionWithoutRun(t *testing.T) {
	pinned := &InstallDeploymentAppBranchRef{ID: "br_1", Name: "continuous-release", RunID: "run_1", CommitSHA: "abc"}
	provision := InstallDeployment{
		Type:      InstallDeploymentTypeProvision,
		AppBranch: &InstallDeploymentAppBranchRef{ID: "br_1", Name: "continuous-release"},
	}
	applyPinnedConfigBranch(&provision, pinned)
	assert.Equal(t, "run_1", provision.AppBranch.RunID)
	assert.Equal(t, "abc", provision.AppBranch.CommitSHA)

	linked := InstallDeployment{
		Type:      InstallDeploymentTypeProvision,
		AppBranch: &InstallDeploymentAppBranchRef{ID: "br_1", Name: "continuous-release", RunID: "run_existing"},
	}
	applyPinnedConfigBranch(&linked, pinned)
	assert.Equal(t, "run_existing", linked.AppBranch.RunID)

	deploy := InstallDeployment{
		Type:      InstallDeploymentTypeComponentDeploy,
		AppBranch: &InstallDeploymentAppBranchRef{ID: "br_1", Name: "continuous-release"},
	}
	applyPinnedConfigBranch(&deploy, pinned)
	assert.Empty(t, deploy.AppBranch.RunID)
}
