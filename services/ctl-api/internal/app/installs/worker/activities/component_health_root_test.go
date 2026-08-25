package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func resourceRow(componentID, kind, namespace, name, health, message string, at time.Time) app.InstallComponentResourceState {
	return app.InstallComponentResourceState{
		InstallComponentID: componentID,
		Source:             app.InstallComponentResourceSourceComponent,
		Provider:           providerKubernetes,
		Kind:               kind,
		Namespace:          namespace,
		Name:               name,
		Health:             health,
		Message:            message,
		ObservedAt:         at,
	}
}

// ClickHouse returns no defined row order, so the named cause must not depend on
// which equally-broken resource arrives first.
func TestRootSelectionIsOrderIndependent(t *testing.T) {
	at := time.Now()
	rows := []app.InstallComponentResourceState{
		resourceRow("c1", "Pod", "ns", "api-abc", "degraded", "pod says so", at),
		resourceRow("c1", "Deployment", "ns", "api", "degraded", "deployment says so", at),
		resourceRow("c1", "ReplicaSet", "ns", "api-1", "degraded", "replicaset says so", at),
	}

	forward := collapseComponentHealthRows(rows)["c1"][0]
	reversed := collapseComponentHealthRows([]app.InstallComponentResourceState{rows[2], rows[1], rows[0]})["c1"][0]

	assert.Equal(t, forward.RootKind, reversed.RootKind)
	assert.Equal(t, forward.RootName, reversed.RootName)
	assert.Equal(t, "Deployment", forward.RootKind,
		"the controller is what the user declared; a pod name changes every rollout")
}

// Severity still wins: a worse pod outranks a merely degraded controller.
func TestSeverityOutranksKind(t *testing.T) {
	at := time.Now()
	got := collapseComponentHealthRows([]app.InstallComponentResourceState{
		resourceRow("c1", "Deployment", "ns", "api", "degraded", "", at),
		resourceRow("c1", "Pod", "ns", "api-abc", "unhealthy", "gone", at),
	})["c1"][0]

	assert.Equal(t, "Pod", got.RootKind)
	assert.Equal(t, app.InstallComponentHealthStatusUnhealthy, got.Health)
}

// Two resources of the same kind and severity resolve by name, not by chance.
func TestEqualCandidatesResolveByName(t *testing.T) {
	at := time.Now()
	got := collapseComponentHealthRows([]app.InstallComponentResourceState{
		resourceRow("c1", "Deployment", "ns", "zebra", "degraded", "", at),
		resourceRow("c1", "Deployment", "ns", "alpha", "degraded", "", at),
	})["c1"][0]

	assert.Equal(t, "alpha", got.RootName)
}

// Naming one resource must not read as one failure.
func TestDescriptionCountsOtherAffected(t *testing.T) {
	now := time.Now()
	rep := &componentHealthReport{
		ObservedAt: now,
		Health:     app.InstallComponentHealthStatusDegraded,
		RootKind:   "Deployment",
		RootName:   "oomkill",
		Message:    "container c: OOMKilled",
		Resources:  30,
		ResourceCounts: map[string]int{
			"healthy":     15,
			"progressing": 9,
			"degraded":    5,
			"unhealthy":   1,
		},
	}

	got := componentHealthDescription(app.InstallComponentHealthStatusDegraded, rep, now)
	assert.Equal(t, "Deployment oomkill: container c: OOMKilled (+5 more affected)", got)
}

// A lone failure says nothing extra, and progressing has no affected count.
func TestDescriptionOmitsCountWhenAlone(t *testing.T) {
	now := time.Now()
	rep := &componentHealthReport{
		ObservedAt:     now,
		Health:         app.InstallComponentHealthStatusDegraded,
		RootKind:       "HorizontalPodAutoscaler",
		RootNamespace:  "whoami",
		RootName:       "whoami",
		Message:        "failed to get cpu utilization",
		Resources:      3,
		ResourceCounts: map[string]int{"healthy": 2, "degraded": 1},
	}

	assert.Equal(t, "HorizontalPodAutoscaler whoami/whoami: failed to get cpu utilization",
		componentHealthDescription(app.InstallComponentHealthStatusDegraded, rep, now))

	rep.Health = app.InstallComponentHealthStatusProgressing
	rep.ResourceCounts = map[string]int{"progressing": 3}
	assert.NotContains(t, componentHealthDescription(app.InstallComponentHealthStatusProgressing, rep, now), "more affected")
}

// A gate-refused deploy is applied and live, so anything asking "what plan is
// running" must still count it; only "did it succeed" must not.
func TestAppliedDeployStatusesIncludeHealthFailed(t *testing.T) {
	statuses := app.AppliedDeployStatuses()

	assert.Contains(t, statuses, app.InstallDeployStatusActive)
	assert.Contains(t, statuses, app.InstallDeployStatusHealthFailed)
	assert.NotContains(t, statuses, app.InstallDeployStatusError,
		"an errored deploy never applied, so its plan is not live")
}

// The component must read as failing when its deploy was refused.
func TestHealthFailedDeployMapsToComponentError(t *testing.T) {
	assert.Equal(t, app.InstallComponentStatusError,
		app.DeployStatusToComponentStatus(app.InstallDeployStatusHealthFailed))
	assert.Equal(t, app.InstallComponentStatusActive,
		app.DeployStatusToComponentStatus(app.InstallDeployStatusActive))
}
