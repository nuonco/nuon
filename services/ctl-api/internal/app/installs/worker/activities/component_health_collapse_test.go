package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func row(kind, name, health string, observedAt time.Time) app.InstallComponentResourceState {
	return app.InstallComponentResourceState{
		InstallComponentID: "ic1",
		Provider:           "kubernetes",
		Kind:               kind,
		Name:               name,
		Namespace:          "whoami",
		Health:             health,
		ObservedAt:         observedAt,
	}
}

// Pins a live failure: an unrunnable probe won the worst-of fold and reported
// unknown while 7 other resources were healthy and observations kept arriving.
func TestCollapseUnknownResourceDoesNotMaskHealthyOnes(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 7, 29, 14, 44, 0, 0, time.UTC)
	rows := []app.InstallComponentResourceState{
		row("HTTPProbe", "unresolvable-target", "unknown", at),
		row("Deployment", "whoami", "healthy", at),
		row("Service", "whoami", "healthy", at),
		row("Ingress", "public", "healthy", at),
	}

	out := collapseComponentHealthRows(rows)
	require.Len(t, out["ic1"], 1)
	rep := out["ic1"][0]

	assert.Equal(t, app.InstallComponentHealthStatusHealthy, rep.Health,
		"one unrunnable probe must not mask three healthy resources")
	assert.Equal(t, 4, rep.Resources, "every resource is still counted")
	assert.Equal(t, 1, rep.ResourceCounts["unknown"], "the unknown row stays visible in the counts")
	assert.Equal(t, 3, rep.ResourceCounts["healthy"])
}

func TestCollapseUnknownLosesRegardlessOfRowOrder(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 7, 29, 14, 44, 0, 0, time.UTC)

	// The unknown row seeded the report in the live case because it sorted first.
	for _, tc := range []struct {
		name string
		rows []app.InstallComponentResourceState
	}{
		{"unknown first", []app.InstallComponentResourceState{
			row("HTTPProbe", "p", "unknown", at), row("Deployment", "d", "healthy", at),
		}},
		{"unknown last", []app.InstallComponentResourceState{
			row("Deployment", "d", "healthy", at), row("HTTPProbe", "p", "unknown", at),
		}},
	} {
		out := collapseComponentHealthRows(tc.rows)
		require.Len(t, out["ic1"], 1, tc.name)
		assert.Equal(t, app.InstallComponentHealthStatusHealthy, out["ic1"][0].Health, tc.name)
	}
}

func TestCollapseRealFailureStillWinsOverUnknown(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 7, 29, 14, 44, 0, 0, time.UTC)
	rows := []app.InstallComponentResourceState{
		row("HTTPProbe", "unresolvable-target", "unknown", at),
		row("Deployment", "whoami", "degraded", at),
		row("Service", "whoami", "healthy", at),
	}

	out := collapseComponentHealthRows(rows)
	require.Len(t, out["ic1"], 1)
	rep := out["ic1"][0]

	assert.Equal(t, app.InstallComponentHealthStatusDegraded, rep.Health,
		"suppressing unknown must not suppress a genuine failure")
	assert.Equal(t, "Deployment", rep.RootKind, "the root must be the real failure, not the unknown probe")
	assert.Equal(t, "whoami", rep.RootName)
}

func TestCollapseAllUnknownIsUnknown(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 7, 29, 14, 44, 0, 0, time.UTC)
	rows := []app.InstallComponentResourceState{
		row("HTTPProbe", "a", "unknown", at),
		row("HTTPProbe", "b", "unknown", at),
	}

	out := collapseComponentHealthRows(rows)
	require.Len(t, out["ic1"], 1)
	rep := out["ic1"][0]

	assert.Equal(t, app.InstallComponentHealthStatusUnknown, rep.Health,
		"a report where nothing could be assessed is genuinely unknown")
	assert.NotEmpty(t, rep.RootKind, "an all-unknown report still names a root resource")
}

func TestCollapseKeepsReportsSeparatePerTimestampNewestFirst(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 7, 29, 14, 43, 0, 0, time.UTC)
	newer := time.Date(2026, 7, 29, 14, 44, 0, 0, time.UTC)
	rows := []app.InstallComponentResourceState{
		row("Deployment", "whoami", "healthy", older),
		row("Deployment", "whoami", "degraded", newer),
	}

	out := collapseComponentHealthRows(rows)
	require.Len(t, out["ic1"], 2)
	assert.Equal(t, newer, out["ic1"][0].ObservedAt, "newest report first")
	assert.Equal(t, app.InstallComponentHealthStatusDegraded, out["ic1"][0].Health)
	assert.Equal(t, app.InstallComponentHealthStatusHealthy, out["ic1"][1].Health)
}

func TestCollapseSkipsIdentityOnlyProviders(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 7, 29, 14, 44, 0, 0, time.UTC)
	tfRow := row("aws_acm_certificate", "arn:aws:acm:...", "unknown", at)
	tfRow.Provider = "aws"

	rows := []app.InstallComponentResourceState{tfRow, row("Deployment", "whoami", "healthy", at)}

	out := collapseComponentHealthRows(rows)
	require.Len(t, out["ic1"], 1)
	rep := out["ic1"][0]

	assert.Equal(t, app.InstallComponentHealthStatusHealthy, rep.Health)
	assert.Equal(t, 1, rep.Resources, "identity-only cloud rows bear no verdict and are not counted")
}

// unknown has distinct causes (runner silence vs. nothing assessable) that
// must not share one message — conflating them sends the reader to the wrong place.
func TestComponentHealthDescriptionDistinguishesUnknownCauses(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 29, 14, 44, 0, 0, time.UTC)

	t.Run("no reports at all", func(t *testing.T) {
		got := componentHealthDescription(app.InstallComponentHealthStatusUnknown, nil, now)
		assert.Equal(t, "no health observations reported", got)
	})

	t.Run("reports went stale", func(t *testing.T) {
		stale := &componentHealthReport{
			ObservedAt: now.Add(-componentHealthStaleAfter - time.Minute),
			Health:     app.InstallComponentHealthStatusHealthy,
			RootKind:   "Deployment",
			RootName:   "whoami",
		}
		got := componentHealthDescription(app.InstallComponentHealthStatusUnknown, stale, now)
		assert.Equal(t, "no recent health observations from the runner", got)
	})

	t.Run("fresh reports but nothing assessable", func(t *testing.T) {
		fresh := &componentHealthReport{
			ObservedAt: now.Add(-30 * time.Second),
			Health:     app.InstallComponentHealthStatusUnknown,
			RootKind:   "HTTPProbe",
			RootName:   "unresolvable-target",
			Message:    "probe target could not be resolved from install state yet",
		}
		got := componentHealthDescription(app.InstallComponentHealthStatusUnknown, fresh, now)
		assert.NotContains(t, got, "no recent health observations",
			"observations arrived 30s ago — blaming the runner is false")
		assert.Contains(t, got, "unresolvable-target", "name what could not be checked")
		assert.Contains(t, got, "could not be resolved")
	})
}

// A healthy verdict must not describe an unassessable resource as healthy —
// live case: 5 healthy + 1 unrunnable probe reported "all 6 resources healthy".
func TestComponentHealthDescriptionSeparatesUncheckedFromHealthy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 29, 15, 38, 0, 0, time.UTC)

	t.Run("some unchecked", func(t *testing.T) {
		latest := &componentHealthReport{
			ObservedAt:     now.Add(-14 * time.Second),
			Health:         app.InstallComponentHealthStatusHealthy,
			Resources:      6,
			ResourceCounts: map[string]int{"healthy": 5, "unknown": 1},
		}
		got := componentHealthDescription(app.InstallComponentHealthStatusHealthy, latest, now)
		assert.Equal(t, "5 of 6 resources healthy, 1 could not be checked", got)
		assert.NotContains(t, got, "all 6", "an unrunnable check is not a passing check")
	})

	t.Run("none unchecked keeps the simple wording", func(t *testing.T) {
		latest := &componentHealthReport{
			ObservedAt:     now.Add(-14 * time.Second),
			Health:         app.InstallComponentHealthStatusHealthy,
			Resources:      6,
			ResourceCounts: map[string]int{"healthy": 6},
		}
		got := componentHealthDescription(app.InstallComponentHealthStatusHealthy, latest, now)
		assert.Equal(t, "all 6 resources healthy", got)
	})

	t.Run("single resource", func(t *testing.T) {
		latest := &componentHealthReport{
			ObservedAt:     now.Add(-14 * time.Second),
			Health:         app.InstallComponentHealthStatusHealthy,
			Resources:      1,
			ResourceCounts: map[string]int{"healthy": 1},
		}
		got := componentHealthDescription(app.InstallComponentHealthStatusHealthy, latest, now)
		assert.Equal(t, "1 resource healthy", got)
	})
}

// Labelling itself is covered by TestMarkDownstreamLabelling; this covers its
// two consequences: the alert drops but the verdict still reports truthfully.
func TestDownstreamSuppressesAlertButKeepsVerdict(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 29, 16, 20, 0, 0, time.UTC)
	latest := &componentHealthReport{
		ObservedAt:     now.Add(-20 * time.Second),
		Health:         app.InstallComponentHealthStatusUnhealthy,
		RootKind:       "Ingress",
		RootName:       "public",
		Message:        "backend not found",
		Resources:      1,
		ResourceCounts: map[string]int{"unhealthy": 1},
	}
	newEval := func(downstreamOf string) *componentEval {
		return &componentEval{
			ic: &app.InstallComponent{
				ID:          "ic-alb",
				ComponentID: "cmp-alb",
				Component:   app.Component{Name: "application_load_balancer"},
			},
			prior:        app.InstallComponentHealthStatusHealthy,
			verdict:      app.InstallComponentHealthStatusUnhealthy,
			latest:       latest,
			downstreamOf: downstreamOf,
		}
	}

	t.Run("root cause alerts", func(t *testing.T) {
		e := newEval("")
		desc := componentHealthDescriptionFor(e, now)
		_, ok := componentHealthNotificationFor(e, desc)
		assert.True(t, ok, "a component that is its own root cause must alert")
		assert.NotContains(t, desc, "downstream of")
	})

	t.Run("downstream is suppressed but still unhealthy", func(t *testing.T) {
		e := newEval("whoami")
		desc := componentHealthDescriptionFor(e, now)

		_, ok := componentHealthNotificationFor(e, desc)
		assert.False(t, ok, "one root cause must produce one alert, not one per dependent")

		assert.Equal(t, app.InstallComponentHealthStatusUnhealthy, e.verdict,
			"suppressing the alert must not soften the verdict — the component really is unhealthy")
		assert.Contains(t, desc, "(downstream of whoami)",
			"the description must point at the thing to actually go fix")
		assert.Contains(t, desc, "backend not found", "the real root resource is still named")
	})
}

// Pins live churn: redeploys flapped Healthy → Not applicable → Progressing
// → Healthy because deploy status left "deployed" while the old workload kept serving.
func TestComponentVerdictDuringRedeploy(t *testing.T) {
	t.Parallel()

	a := &Activities{}
	now := time.Date(2026, 7, 30, 6, 30, 0, 0, time.UTC)
	healthyReports := []componentHealthReport{{
		ObservedAt: now.Add(-30 * time.Second),
		Health:     app.InstallComponentHealthStatusHealthy,
	}}
	helm := app.Component{Type: app.ComponentTypeHelmChart}

	t.Run("mid-redeploy keeps evaluating", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:       app.InstallComponentStatusExecuting,
			HealthStatus: app.InstallComponentHealthStatusHealthy,
			Component:    helm,
		}
		got := a.componentVerdict(ic, healthyReports, now)
		assert.NotEqual(t, app.InstallComponentHealthStatusNotApplicable, got,
			"the previous workload is still serving — a redeploy must not flap the verdict to not-applicable")
	})

	t.Run("first deploy in flight stays not-applicable", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:       app.InstallComponentStatusExecuting,
			HealthStatus: app.InstallComponentHealthStatusNotApplicable,
			Component:    helm,
		}
		got := a.componentVerdict(ic, healthyReports, now)
		assert.Equal(t, app.InstallComponentHealthStatusNotApplicable, got,
			"no deploy has ever completed — there is no workload a verdict could describe")
	})

	t.Run("disabled always not-applicable, even with a prior verdict", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:       app.InstallComponentStatusDisabled,
			HealthStatus: app.InstallComponentHealthStatusHealthy,
			Component:    helm,
		}
		got := a.componentVerdict(ic, healthyReports, now)
		assert.Equal(t, app.InstallComponentHealthStatusNotApplicable, got)
	})

	t.Run("torn down goes not-applicable even with a prior verdict", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:       app.InstallComponentStatusInactive,
			HealthStatus: app.InstallComponentHealthStatusHealthy,
			Component:    helm,
		}
		got := a.componentVerdict(ic, healthyReports, now)
		assert.Equal(t, app.InstallComponentHealthStatusNotApplicable, got,
			"the workload was deliberately deleted — health must not alert on a removal the user asked for")
	})

	t.Run("failed teardown keeps evaluating", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:       app.InstallComponentStatusDeleteFailed,
			HealthStatus: app.InstallComponentHealthStatusHealthy,
			Component:    helm,
		}
		got := a.componentVerdict(ic, healthyReports, now)
		assert.NotEqual(t, app.InstallComponentHealthStatusNotApplicable, got,
			"resources may linger after a failed teardown — their health is exactly what you want to see")
	})

	t.Run("redeploy of a component that was unknown keeps evaluating", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:       app.InstallComponentStatusSyncing,
			HealthStatus: app.InstallComponentHealthStatusUnknown,
			Component:    helm,
		}
		got := a.componentVerdict(ic, healthyReports, now)
		assert.NotEqual(t, app.InstallComponentHealthStatusNotApplicable, got,
			"unknown is a real verdict — proof a deploy ran before")
	})
}

// Pins a live failure: a runner lost cluster access, so k8s observations
// stopped while an exec probe kept passing — healthy for 12min despite ImagePullBackOff.
func TestClusterBlindComponentGoesUnknown(t *testing.T) {
	t.Parallel()

	a := &Activities{}
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	helm := app.Component{Type: app.ComponentTypeHelmChart}
	probeOnly := []componentHealthReport{{
		ObservedAt:      now.Add(-20 * time.Second),
		Health:          app.InstallComponentHealthStatusHealthy,
		ClusterEvidence: false,
	}}
	withCluster := []componentHealthReport{{
		ObservedAt:      now.Add(-20 * time.Second),
		Health:          app.InstallComponentHealthStatusHealthy,
		ClusterEvidence: true,
	}}
	seen := func() app.CompositeStatus {
		return app.CompositeStatus{Metadata: map[string]any{"cluster_seen": true}}
	}

	t.Run("probe-only evidence cannot hold a healthy verdict", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:         app.InstallComponentStatusActive,
			HealthStatus:   app.InstallComponentHealthStatusHealthy,
			HealthStatusV2: seen(),
			Component:      helm,
		}
		assert.Equal(t, app.InstallComponentHealthStatusUnknown,
			a.componentVerdict(ic, probeOnly, now),
			"a passing probe must not certify a workload nobody can see")
	})

	t.Run("cluster evidence keeps the verdict", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:         app.InstallComponentStatusActive,
			HealthStatus:   app.InstallComponentHealthStatusHealthy,
			HealthStatusV2: seen(),
			Component:      helm,
		}
		assert.Equal(t, app.InstallComponentHealthStatusHealthy,
			a.componentVerdict(ic, withCluster, now))
	})

	t.Run("a component that never had cluster evidence stays probe-assessable", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:       app.InstallComponentStatusActive,
			HealthStatus: app.InstallComponentHealthStatusHealthy,
			Component:    helm,
		}
		assert.Equal(t, app.InstallComponentHealthStatusHealthy,
			a.componentVerdict(ic, probeOnly, now),
			"a chart owning no watched kinds must not be pinned to unknown forever")
	})

	t.Run("terraform components are unaffected", func(t *testing.T) {
		ic := &app.InstallComponent{
			Status:         app.InstallComponentStatusActive,
			HealthStatus:   app.InstallComponentHealthStatusHealthy,
			HealthStatusV2: seen(),
			Component:      app.Component{Type: app.ComponentTypeTerraformModule},
		}
		assert.Equal(t, app.InstallComponentHealthStatusHealthy,
			a.componentVerdict(ic, probeOnly, now),
			"probes are the primary evidence for a terraform module, not a fallback")
	})

	t.Run("the description names the cause", func(t *testing.T) {
		e := &componentEval{
			ic:           &app.InstallComponent{Component: helm},
			verdict:      app.InstallComponentHealthStatusUnknown,
			clusterBlind: true,
		}
		desc := componentHealthDescriptionFor(e, now)
		assert.Contains(t, desc, "cluster resources")
		assert.NotContains(t, desc, "no recent health observations",
			"blaming runner silence would send the reader to the wrong thing while probes are arriving")
	})
}

// Cluster evidence must go stale on the same clock the resource rows do, or the
// table shows every row stale while the verdict still says healthy.
func TestClusterEvidenceMustBeFresh(t *testing.T) {
	t.Parallel()

	a := &Activities{}
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	ic := func() *app.InstallComponent {
		return &app.InstallComponent{
			Status:         app.InstallComponentStatusActive,
			HealthStatus:   app.InstallComponentHealthStatusHealthy,
			HealthStatusV2: app.CompositeStatus{Metadata: map[string]any{"cluster_seen": true}},
			Component:      app.Component{Type: app.ComponentTypeHelmChart},
		}
	}
	reports := func(clusterAge time.Duration) []componentHealthReport {
		return []componentHealthReport{
			{ObservedAt: now.Add(-10 * time.Second), Health: app.InstallComponentHealthStatusHealthy},
			{ObservedAt: now.Add(-clusterAge), Health: app.InstallComponentHealthStatusHealthy, ClusterEvidence: true},
		}
	}

	assert.Equal(t, app.InstallComponentHealthStatusHealthy,
		a.componentVerdict(ic(), reports(2*time.Minute), now),
		"cluster evidence inside the staleness threshold still certifies")

	assert.Equal(t, app.InstallComponentHealthStatusUnknown,
		a.componentVerdict(ic(), reports(9*time.Minute), now),
		"cluster evidence the resource table would mark stale must not certify")
}

// A pushed check reports on its own cadence; before TTLs it kept voting its
// last verdict until aging out of the read window, then silently vanished.
func TestCustomCheckStaleAfter(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	rows := func(staleAfter uint32, pushedAgo time.Duration) []app.InstallComponentResourceState {
		return []app.InstallComponentResourceState{
			{InstallComponentID: "ic1", Provider: providerKubernetes, Kind: "Deployment", Name: "api",
				Health: "healthy", ObservedAt: now},
			{InstallComponentID: "ic1", Provider: providerCustom, Kind: "CustomCheck", Name: "nightly-audit",
				Health: "unhealthy", ObservedAt: now.Add(-pushedAgo), StaleAfterSeconds: staleAfter},
		}
	}

	t.Run("inside its declared ttl the check still votes", func(t *testing.T) {
		reports := collapseComponentHealthRows(rows(1800, 20*time.Minute))["ic1"]
		require.NotEmpty(t, reports)
		assert.Equal(t, app.InstallComponentHealthStatusUnhealthy, reports[0].Health,
			"a 30m ttl must survive a 20m-old report")
	})

	t.Run("past its ttl it reads unknown and is still counted", func(t *testing.T) {
		reports := collapseComponentHealthRows(rows(600, 20*time.Minute))["ic1"]
		require.NotEmpty(t, reports)
		assert.Equal(t, app.InstallComponentHealthStatusHealthy, reports[0].Health,
			"an expired check must not keep voting unhealthy")
		assert.Equal(t, 1, reports[0].ResourceCounts[string(app.InstallComponentHealthStatusUnknown)],
			"it must still be counted as unassessable, not dropped")
	})

	t.Run("no ttl declared falls back to the default", func(t *testing.T) {
		fresh := collapseComponentHealthRows(rows(0, 2*time.Minute))["ic1"]
		require.NotEmpty(t, fresh)
		assert.Equal(t, app.InstallComponentHealthStatusUnhealthy, fresh[0].Health)

		expired := collapseComponentHealthRows(rows(0, 9*time.Minute))["ic1"]
		require.NotEmpty(t, expired)
		assert.Equal(t, app.InstallComponentHealthStatusHealthy, expired[0].Health,
			"default ttl is the same 5m the UI draws stale at")
	})
}

// unknown from a custom check means nobody could assess it, not a severity —
// it must not outrank resources that did report, same rule as the cluster fold.
func TestUnknownCustomCheckDoesNotMaskHealthy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)

	t.Run("unknown check alongside a healthy deployment", func(t *testing.T) {
		reports := collapseComponentHealthRows([]app.InstallComponentResourceState{
			{InstallComponentID: "ic1", Provider: providerKubernetes, Kind: "Deployment", Name: "api",
				Health: "healthy", ObservedAt: now},
			{InstallComponentID: "ic1", Provider: providerCustom, Kind: "CustomCheck", Name: "audit",
				Health: "unknown", ObservedAt: now},
		})["ic1"]
		require.NotEmpty(t, reports)
		assert.Equal(t, app.InstallComponentHealthStatusHealthy, reports[0].Health)
		assert.Equal(t, 2, reports[0].Resources, "the unknown check is still counted")
	})

	t.Run("unknown check alone is unknown, not healthy", func(t *testing.T) {
		reports := collapseComponentHealthRows([]app.InstallComponentResourceState{
			{InstallComponentID: "ic1", Provider: providerCustom, Kind: "CustomCheck", Name: "audit",
				Health: "unknown", ObservedAt: now},
		})["ic1"]
		require.NotEmpty(t, reports)
		assert.Equal(t, app.InstallComponentHealthStatusUnknown, reports[0].Health,
			"a component whose only check could not be assessed is unknown")
	})

	t.Run("an unhealthy check still wins", func(t *testing.T) {
		reports := collapseComponentHealthRows([]app.InstallComponentResourceState{
			{InstallComponentID: "ic1", Provider: providerKubernetes, Kind: "Deployment", Name: "api",
				Health: "healthy", ObservedAt: now},
			{InstallComponentID: "ic1", Provider: providerCustom, Kind: "CustomCheck", Name: "audit",
				Health: "unhealthy", Message: "audit failed", ObservedAt: now},
		})["ic1"]
		require.NotEmpty(t, reports)
		assert.Equal(t, app.InstallComponentHealthStatusUnhealthy, reports[0].Health)
		assert.Equal(t, "audit", reports[0].RootName)
	})
}
