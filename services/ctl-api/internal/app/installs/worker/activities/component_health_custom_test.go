package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBearsVerdict(t *testing.T) {
	// Terraform enumeration is inventory, not assessment — counting it would
	// drag every terraform component to unknown.
	assert.False(t, bearsVerdict("aws"))
	assert.False(t, bearsVerdict("gcp"))
	assert.False(t, bearsVerdict("azure"))

	assert.True(t, bearsVerdict("kubernetes"))
	assert.True(t, bearsVerdict("probe"))
	assert.True(t, bearsVerdict("custom"))
}

func TestApplyCustomChecks(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	healthy := app.InstallComponentHealthStatusHealthy
	degraded := app.InstallComponentHealthStatusDegraded
	unhealthy := app.InstallComponentHealthStatusUnhealthy

	// Runner reports, newest first, as recentComponentHealthReports produces.
	runnerReports := func(healths ...app.InstallComponentHealthStatus) []componentHealthReport {
		out := make([]componentHealthReport, 0, len(healths))
		for i, h := range healths {
			out = append(out, componentHealthReport{
				ObservedAt:     now.Add(-time.Duration(i) * time.Minute),
				Health:         h,
				RootKind:       "Deployment",
				RootName:       "api",
				Resources:      1,
				ResourceCounts: map[string]int{string(h): 1},
			})
		}
		return out
	}

	t.Run("no custom checks leaves reports untouched", func(t *testing.T) {
		reports := applyCustomChecks(runnerReports(healthy, healthy), nil)
		require.Len(t, reports, 2)
		assert.Equal(t, healthy, reports[0].Health)
	})

	t.Run("an unhealthy custom check worsens every report it was in effect for", func(t *testing.T) {
		reports := applyCustomChecks(runnerReports(healthy, healthy, healthy), []customCheckObservation{
			{Name: "checkout-latency", Health: unhealthy, Message: "p99 4.2s", ObservedAt: now.Add(-3 * time.Minute)},
		})

		require.Len(t, reports, 3)
		// Every report is at or after the check, so all three see it — which is
		// what lets the existing 3-bad debounce apply to custom checks.
		for i := range reports {
			assert.Equal(t, unhealthy, reports[i].Health, "report %d", i)
			assert.Equal(t, "CustomCheck", reports[i].RootKind)
			assert.Equal(t, "checkout-latency", reports[i].RootName)
			assert.Equal(t, "p99 4.2s", reports[i].Message)
		}
	})

	t.Run("a check reported after an older report does not backdate onto it", func(t *testing.T) {
		// Check landed 1 minute ago; the report from 2 minutes ago predates it.
		reports := applyCustomChecks(runnerReports(healthy, healthy, healthy), []customCheckObservation{
			{Name: "queue-depth", Health: unhealthy, ObservedAt: now.Add(-time.Minute)},
		})

		require.Len(t, reports, 3)
		assert.Equal(t, unhealthy, reports[0].Health, "newest report is after the check")
		assert.Equal(t, unhealthy, reports[1].Health, "report at the check time sees it")
		assert.Equal(t, healthy, reports[2].Health, "report predating the check is untouched")
	})

	t.Run("a healthy custom check never improves a failing kubernetes verdict", func(t *testing.T) {
		reports := applyCustomChecks(runnerReports(unhealthy), []customCheckObservation{
			{Name: "synthetic", Health: healthy, ObservedAt: now.Add(-time.Minute)},
		})

		require.Len(t, reports, 1)
		assert.Equal(t, unhealthy, reports[0].Health)
		assert.Equal(t, "Deployment", reports[0].RootKind, "kubernetes stays the root cause")
	})

	t.Run("the worst of several checks wins", func(t *testing.T) {
		reports := applyCustomChecks(runnerReports(healthy), []customCheckObservation{
			{Name: "a", Health: degraded, ObservedAt: now.Add(-2 * time.Minute)},
			{Name: "b", Health: unhealthy, Message: "b is down", ObservedAt: now.Add(-2 * time.Minute)},
		})

		require.Len(t, reports, 1)
		assert.Equal(t, unhealthy, reports[0].Health)
		assert.Equal(t, "b", reports[0].RootName)
	})

	t.Run("only the latest state of a check counts", func(t *testing.T) {
		reports := applyCustomChecks(runnerReports(healthy), []customCheckObservation{
			{Name: "flaky", Health: unhealthy, ObservedAt: now.Add(-5 * time.Minute)},
			{Name: "flaky", Health: healthy, ObservedAt: now.Add(-time.Minute)},
		})

		require.Len(t, reports, 1)
		assert.Equal(t, healthy, reports[0].Health, "the recovery supersedes the earlier failure")
	})

	t.Run("custom checks alone become the timeline when nothing else reports", func(t *testing.T) {
		reports := applyCustomChecks(nil, []customCheckObservation{
			{Name: "rds-latency", Health: unhealthy, Message: "slow", ObservedAt: now.Add(-2 * time.Minute)},
			{Name: "rds-latency", Health: unhealthy, Message: "slow", ObservedAt: now.Add(-time.Minute)},
		})

		require.Len(t, reports, 2, "a component with only custom checks is still observable")
		assert.Equal(t, unhealthy, reports[0].Health)
		assert.Equal(t, "rds-latency", reports[0].RootName)
		assert.True(t, reports[0].ObservedAt.After(reports[1].ObservedAt), "newest first")
	})
}
