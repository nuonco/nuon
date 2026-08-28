package customermanaged

import (
	"testing"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestAggregateHealth(t *testing.T) {
	tests := []struct {
		name      string
		resources []ResourceHealth
		want      string
	}{
		{name: "empty", want: "unknown"},
		{name: "all healthy", resources: []ResourceHealth{{Health: "healthy"}, {Health: "healthy"}}, want: "healthy"},
		{name: "worst wins", resources: []ResourceHealth{{Health: "healthy"}, {Health: "degraded"}, {Health: "progressing"}}, want: "degraded"},
		{name: "unhealthy beats degraded", resources: []ResourceHealth{{Health: "degraded"}, {Health: "unhealthy"}}, want: "unhealthy"},
		{name: "not-applicable ignored", resources: []ResourceHealth{{Health: "not-applicable"}, {Health: "healthy"}}, want: "healthy"},
		{name: "only unscored", resources: []ResourceHealth{{Health: "not-applicable"}, {Health: ""}}, want: "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aggregateHealth(tt.resources); got != tt.want {
				t.Fatalf("aggregateHealth = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewHealthSnapshotResolvesNamesAndAggregates(t *testing.T) {
	installComponentID := "inc-1"
	specs := []ComponentSpec{
		{InstallComponentID: "inc-1", ComponentID: "cmp-1", ComponentName: "api", ComponentType: "helm_chart"},
		{InstallComponentID: "inc-2", ComponentID: "cmp-2", ComponentName: "worker", ComponentType: "helm_chart"},
	}
	req := &models.ServiceCreateComponentHealthRequest{
		Kind:       "watch",
		ObservedAt: "2026-08-10T10:00:00Z",
		Components: []*models.ServiceComponentHealthComponent{
			{
				ComponentID:        "cmp-1",
				InstallComponentID: &installComponentID,
				Resources: []*models.ServiceComponentHealthResource{
					{Kind: "Deployment", Name: "api", Health: "healthy"},
					{Kind: "Pod", Name: "api-0", Health: "degraded", Message: "CrashLoopBackOff"},
				},
			},
			nil,
		},
		SandboxReleases: []*models.ServiceComponentHealthSandboxRelease{
			{ReleaseName: strPtr("ingress"), Namespace: "kube-system", Resources: []*models.ServiceComponentHealthResource{{Kind: "DaemonSet", Health: "healthy"}}},
		},
	}

	snapshot := NewHealthSnapshot(req, specs)
	if !snapshot.ObservedAt.Equal(time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("observed_at not parsed: %s", snapshot.ObservedAt)
	}
	if len(snapshot.Components) != 1 {
		t.Fatalf("nil components should be skipped, got %d", len(snapshot.Components))
	}
	c := snapshot.Components[0]
	if c.ComponentName != "api" || c.ComponentType != "helm_chart" {
		t.Fatalf("spec metadata not resolved: %+v", c)
	}
	if c.Health != "degraded" {
		t.Fatalf("aggregate health = %q, want degraded", c.Health)
	}
	if len(snapshot.SandboxReleases) != 1 || snapshot.SandboxReleases[0].ReleaseName != "ingress" || snapshot.SandboxReleases[0].Health != "healthy" {
		t.Fatalf("unexpected sandbox releases: %+v", snapshot.SandboxReleases)
	}
}

func TestHealthTransitions(t *testing.T) {
	now := time.Now().UTC()
	current := &HealthSnapshot{
		ObservedAt: now,
		Components: []ComponentHealth{
			{ComponentID: "cmp-1", ComponentName: "api", Health: "degraded"},
			{ComponentID: "cmp-2", ComponentName: "worker", Health: "healthy"},
		},
		SandboxReleases: []SandboxReleaseHealth{{ReleaseName: "ingress", Health: "healthy"}},
	}

	first := HealthTransitions(nil, current)
	if len(first) != 3 {
		t.Fatalf("first observation should record every component, got %d", len(first))
	}
	if first[0].From != "" || first[0].To != "degraded" {
		t.Fatalf("first transition should come from empty state: %+v", first[0])
	}

	previous := &HealthSnapshot{
		Components: []ComponentHealth{
			{ComponentID: "cmp-1", ComponentName: "api", Health: "healthy"},
			{ComponentID: "cmp-2", ComponentName: "worker", Health: "healthy"},
		},
		SandboxReleases: []SandboxReleaseHealth{{ReleaseName: "ingress", Health: "healthy"}},
	}
	changed := HealthTransitions(previous, current)
	if len(changed) != 1 {
		t.Fatalf("only cmp-1 changed, got %d transitions", len(changed))
	}
	if changed[0].ComponentID != "cmp-1" || changed[0].From != "healthy" || changed[0].To != "degraded" {
		t.Fatalf("unexpected transition: %+v", changed[0])
	}

	if same := HealthTransitions(current, current); len(same) != 0 {
		t.Fatalf("identical snapshots should not transition, got %+v", same)
	}
}

func strPtr(s string) *string { return &s }
