package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationstate"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

func TestPortalCombinesInstallationUpgradeAndOperationRuns(t *testing.T) {
	stateDir := t.TempDir()
	deploymentDir := t.TempDir()
	initialStarted := time.Date(2026, 8, 19, 10, 20, 0, 0, time.UTC)
	initialFinished := initialStarted.Add(30 * time.Minute)
	upgradeStarted := initialFinished.Add(10 * time.Minute)
	upgradeFinished := upgradeStarted.Add(2 * time.Minute)
	operationStarted := upgradeFinished.Add(time.Minute)

	writeStateObject(t, deploymentDir, "state/status.json", statestore.Status{
		InstallID: "install-1", BundleDigest: "sha256:v1", RunID: "install-run", Status: statestore.RunStatusFinished,
		StartedAt: initialStarted, FinishedAt: &initialFinished,
		Steps: []statestore.StepStatus{{ID: "deploy-api", Name: "api apply", Status: "finished", StartedAt: &initialStarted, FinishedAt: &initialFinished}},
	})
	writeStateObject(t, stateDir, "status.json", statestore.Status{
		InstallID: "install-1", BundleDigest: "sha256:v2", RunID: "upgrade-run", Status: statestore.RunStatusFinished,
		StartedAt: upgradeStarted, FinishedAt: &upgradeFinished,
		Steps: []statestore.StepStatus{
			{ID: "deploy-api", Name: "api apply", Status: "auto-skipped", SourceRunID: "install-run"},
			{ID: "deploy-worker", Name: "worker apply", Status: "finished", StartedAt: &upgradeStarted, FinishedAt: &upgradeFinished},
		},
	})
	writeStateObject(t, stateDir, operation.RunStatusKey("operation-run"), operation.RunStatus{
		RunID: "operation-run", RefID: "restart", RefKind: operation.RefKindAction, RefName: "Restart", Source: operation.SourcePortal,
		Status: operation.RunStatusFinished, StartedAt: operationStarted,
	})

	portal, err := newPortalServer(operationstate.NewLocal(stateDir), operationstate.NewLocal(deploymentDir), "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	runs, err := portal.listRuns(httptest.NewRequest("GET", "http://127.0.0.1:1234/api/runs", nil))
	require.NoError(t, err)
	require.Len(t, runs, 3)
	require.Equal(t, "operation-run", runs[0].RunID)
	require.Equal(t, "upgrade-run", runs[1].RunID)
	require.Equal(t, statestore.RunTypeUpgrade, runs[1].RefKind)
	require.Equal(t, "install-run", runs[1].PreviousRunID)
	require.Equal(t, "auto-skipped", runs[1].Steps[0].Status)
	require.Equal(t, "install-run", runs[1].Steps[0].SourceRunID)
	require.Empty(t, runs[1].Steps[0].JobID)
	require.Equal(t, "install-run", runs[2].RunID)
	require.Equal(t, statestore.RunTypeInstall, runs[2].RefKind)
}

func TestBootstrapPortalRunDoesNotInferLegacyStepSuccess(t *testing.T) {
	started := time.Date(2026, 8, 19, 10, 20, 0, 0, time.UTC)
	run := bootstrapPortalRun(statestore.Status{
		RunID: "upgrade-run", RunType: statestore.RunTypeUpgrade, Status: statestore.RunStatusFinished,
		StartedAt: started, Steps: []statestore.StepStatus{{ID: "sandbox", Name: "sandbox", Status: "finished"}},
	}, "", "failed-run")

	require.Equal(t, "unknown", run.Steps[0].Status)
	require.Equal(t, "Prior execution record unavailable", run.Steps[0].Description)
	require.Empty(t, run.Steps[0].SourceRunID)
}
