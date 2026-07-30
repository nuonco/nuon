package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRetainDeployedComponentResources(t *testing.T) {
	t.Parallel()

	resources := []app.InstallComponentResourceState{
		{Name: "deployed-workload", InstallComponentID: "ic-deployed", Source: app.InstallComponentResourceSourceComponent},
		{Name: "stale-probe", InstallComponentID: "ic-never-deployed", Source: app.InstallComponentResourceSourceComponent},
		{Name: "stale-custom-check", InstallComponentID: "ic-never-deployed", Source: app.InstallComponentResourceSourceComponent},
		{Name: "orphaned-row", InstallComponentID: "ic-gone", Source: app.InstallComponentResourceSourceComponent},
		{Name: "cert-manager", InstallComponentID: "", Source: app.InstallComponentResourceSourceSandbox},
	}

	deployed := map[string]bool{
		"ic-deployed":       true,
		"ic-never-deployed": false,
	}

	kept := retainDeployedComponentResources(resources, deployed)

	names := make([]string, 0, len(kept))
	for _, r := range kept {
		names = append(names, r.Name)
	}

	assert.Equal(t, []string{"deployed-workload", "cert-manager"}, names,
		"only rows from deployed components and sandbox rows survive")
}

func TestRetainDeployedComponentResourcesKeepsSandboxWhenNoComponents(t *testing.T) {
	t.Parallel()

	resources := []app.InstallComponentResourceState{
		{Name: "ebs-csi-node", Source: app.InstallComponentResourceSourceSandbox},
	}

	kept := retainDeployedComponentResources(resources, map[string]bool{})

	assert.Len(t, kept, 1, "sandbox resources are not governed by component deploy state")
}

// A redeploy must not blank the component's rows for the duration of the
// deploy; a teardown must hide them.
func TestEverDeployedForResourceVisibility(t *testing.T) {
	t.Parallel()

	redeploying := app.InstallComponent{
		Status:       app.InstallComponentStatusExecuting,
		HealthStatus: app.InstallComponentHealthStatusHealthy,
	}
	assert.True(t, redeploying.EverDeployed(), "mid-redeploy rows stay visible")

	tornDown := app.InstallComponent{
		Status:       app.InstallComponentStatusInactive,
		HealthStatus: app.InstallComponentHealthStatusHealthy,
	}
	assert.False(t, tornDown.EverDeployed(), "torn-down rows are hidden")

	neverDeployed := app.InstallComponent{
		Status:       app.InstallComponentStatusQueued,
		HealthStatus: app.InstallComponentHealthStatusNotApplicable,
	}
	assert.False(t, neverDeployed.EverDeployed(), "first deploy in flight stays hidden")
}

func TestInstallComponentStatusHasDeployed(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		status app.InstallComponentStatus
		want   bool
	}{
		{app.InstallComponentStatusActive, true},
		{app.InstallComponentStatusNoop, true},
		{app.InstallComponentStatusError, true},
		{app.InstallComponentStatusDisabled, false},
		{app.InstallComponentStatus(""), false},
	} {
		assert.Equal(t, tc.want, tc.status.HasDeployed(), "status %q", tc.status)
	}
}

// Observations outlive config: a probe deleted from the component's config
// keeps its last ClickHouse row for days. Those rows must be labelled so a
// deleted probe's final reading cannot pass for a live check.
func TestMarkRemovedProbes(t *testing.T) {
	t.Parallel()

	rows := []app.InstallComponentResourceState{
		{Source: "component", InstallComponentID: "ic1", Kind: "ExecProbe", Name: "gate-test-always-fails"},
		{Source: "component", InstallComponentID: "ic1", Kind: "HTTPProbe", Name: "public-endpoint"},
		{Source: "component", InstallComponentID: "ic1", Kind: "ExecProbe", Name: "always-ok"},
		{Source: "component", InstallComponentID: "ic1", Kind: "Deployment", Name: "whoami"},
		{Source: "component", InstallComponentID: "ic1", Kind: "CustomCheck", Name: "checkout-latency"},
		{Source: "sandbox", InstallComponentID: "", Kind: "HTTPProbe", Name: "whatever"},
	}
	declared := map[string]map[string]bool{
		"ic1": {"always-ok": true},
	}

	markRemovedProbes(rows, declared)

	byName := map[string]bool{}
	for _, r := range rows {
		byName[r.Name] = r.RemovedFromConfig
	}

	assert.True(t, byName["gate-test-always-fails"], "deleted probe is marked removed")
	assert.True(t, byName["public-endpoint"], "probe moved to another component is removed here")
	assert.False(t, byName["always-ok"], "declared probe is live")
	assert.False(t, byName["whoami"], "k8s resources are not probes and never marked")
	assert.False(t, byName["checkout-latency"], "custom checks are pushed, not declared — cannot be classified removed")
	assert.False(t, byName["whatever"], "sandbox rows are not governed by component config")
}

func TestMarkRemovedProbesUnknownComponentUntouched(t *testing.T) {
	t.Parallel()

	rows := []app.InstallComponentResourceState{
		{Source: "component", InstallComponentID: "ic-gone", Kind: "ExecProbe", Name: "p"},
	}
	markRemovedProbes(rows, map[string]map[string]bool{})
	assert.False(t, rows[0].RemovedFromConfig,
		"a component we could not resolve config for is left unlabelled rather than guessed at")
}
