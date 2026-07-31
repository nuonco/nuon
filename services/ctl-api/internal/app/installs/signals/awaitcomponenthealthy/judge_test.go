package awaitcomponenthealthy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func report(observedAt time.Time, health, rootKind, rootName, message string) activities.GateHealthReport {
	return activities.GateHealthReport{
		ObservedAtTS: observedAt.Unix(),
		Health:       health,
		RootKind:     rootKind,
		RootName:     rootName,
		Message:      message,
	}
}

func TestJudgeWindow(t *testing.T) {
	t.Parallel()

	gateAt := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	window := time.Minute

	t.Run("bad observation inside the window fails immediately", func(t *testing.T) {
		reports := []activities.GateHealthReport{
			report(gateAt.Add(30*time.Second), "unhealthy", "ExecProbe", "gate-test-always-fails", "exit code 1"),
			report(gateAt.Add(7*time.Second), "healthy", "Deployment", "whoami", ""),
		}
		outcome, worst := judgeWindow(reports, gateAt, gateAt.Add(35*time.Second), window, false)
		assert.Equal(t, windowFailBad, outcome, "post-apply bad evidence must not wait for the window")
		require.NotNil(t, worst)
		assert.Equal(t, "gate-test-always-fails", worst.RootName)
	})

	t.Run("pre-apply observations are ignored", func(t *testing.T) {
		reports := []activities.GateHealthReport{
			report(gateAt.Add(-30*time.Second), "unhealthy", "ExecProbe", "old-failure", "stale"),
		}
		outcome, _ := judgeWindow(reports, gateAt, gateAt.Add(10*time.Second), window, false)
		assert.Equal(t, windowWait, outcome,
			"the previous deploy's failures are not evidence about this one")
	})

	t.Run("all healthy passes exactly at the window boundary", func(t *testing.T) {
		reports := []activities.GateHealthReport{
			report(gateAt.Add(55*time.Second), "healthy", "Deployment", "whoami", ""),
			report(gateAt.Add(7*time.Second), "healthy", "Deployment", "whoami", ""),
		}
		outcome, _ := judgeWindow(reports, gateAt, gateAt.Add(window), window, false)
		assert.Equal(t, windowPass, outcome, "no grace: the gate concludes when the window does")
	})

	t.Run("still progressing when the window ends fails", func(t *testing.T) {
		reports := []activities.GateHealthReport{
			report(gateAt.Add(50*time.Second), "progressing", "Deployment", "whoami", "waiting for rollout"),
		}
		outcome, latest := judgeWindow(reports, gateAt, gateAt.Add(window), window, false)
		assert.Equal(t, windowFailState, outcome, "progressing at close is not held-healthy")
		require.NotNil(t, latest)
		assert.Equal(t, "progressing", latest.Health)
	})

	t.Run("no in-window reports asks for a bounded extension", func(t *testing.T) {
		outcome, _ := judgeWindow(nil, gateAt, gateAt.Add(window), window, false)
		assert.Equal(t, windowNeedData, outcome,
			"absence of data extends briefly for the next report instead of failing a good deploy")
	})

	t.Run("healthy but window still open keeps waiting", func(t *testing.T) {
		reports := []activities.GateHealthReport{
			report(gateAt.Add(7*time.Second), "healthy", "Deployment", "whoami", ""),
		}
		outcome, _ := judgeWindow(reports, gateAt, gateAt.Add(30*time.Second), window, false)
		assert.Equal(t, windowWait, outcome, "one healthy report does not shortcut the hold")
	})
}

// Regression: a retry re-applies with no fresh observations mid-flight, and
// that must read as "wait", never a pass.
func TestGateIsNotBypassedByItsOwnRetry(t *testing.T) {
	t.Parallel()

	gateAt := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	window := time.Minute

	outcome, _ := judgeWindow(nil, gateAt, gateAt.Add(10*time.Second), window, false)
	assert.Equal(t, windowWait, outcome, "no data mid-window is a wait, not a free pass")

	outcome, _ = judgeWindow(nil, gateAt, gateAt.Add(window), window, false)
	assert.NotEqual(t, windowPass, outcome, "no data at the window end must never pass outright")
}

func TestWindowNarration(t *testing.T) {
	t.Parallel()

	gateAt := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	window := time.Minute
	healthy := report(gateAt.Add(7*time.Second), "healthy", "Deployment", "whoami", "")

	got := windowNarration(windowWait, &healthy, gateAt, gateAt.Add(25*time.Second), window, nil)
	assert.Equal(t, "healthy so far — 35s of the 1m0s window left — healthy (Deployment whoami)", got)

	got = windowNarration(windowWait, nil, gateAt, gateAt.Add(10*time.Second), window, nil)
	assert.Contains(t, got, "waiting for the first health report")

	got = windowNarration(windowPass, &healthy, gateAt, gateAt.Add(window), window,
		[]activities.ComponentHealthCheckRow{{Name: "cert", Health: "healthy"}})
	assert.Contains(t, got, "held healthy for 1m0s")

	bad := report(gateAt.Add(30*time.Second), "unhealthy", "ExecProbe", "gate-test-always-fails", "exit code 1")
	got = windowNarration(windowFailBad, &bad, gateAt, gateAt.Add(31*time.Second), window, nil)
	assert.Equal(t, "component is unhealthy: ExecProbe gate-test-always-fails: exit code 1", got)
}

// Pins a real gap: on first deploy the runner picks up probes one report
// cycle late, so a window that closes there must not pass on checks that never ran.
func TestMissingProbes(t *testing.T) {
	t.Parallel()

	gateAt := time.Date(2026, 7, 30, 5, 37, 0, 0, time.UTC)
	declared := []string{"public-endpoint", "always-ok", "gate-test-always-fails"}

	check := func(kind, name string, at time.Time) activities.ComponentHealthCheckRow {
		return activities.ComponentHealthCheckRow{Kind: kind, Name: name, Health: "healthy", ObservedAtTS: at.Unix()}
	}

	t.Run("probe-less first report leaves all probes missing", func(t *testing.T) {
		checks := []activities.ComponentHealthCheckRow{
			check("Deployment", "whoami", gateAt.Add(20*time.Second)),
			check("Service", "whoami", gateAt.Add(20*time.Second)),
		}
		missing := missingProbes(declared, checks, gateAt)
		assert.Equal(t, declared, missing, "k8s rows alone must not satisfy probe coverage")
	})

	t.Run("stale probe rows from before the gate do not count", func(t *testing.T) {
		checks := []activities.ComponentHealthCheckRow{
			check("ExecProbe", "always-ok", gateAt.Add(-5*time.Minute)),
		}
		missing := missingProbes(declared, checks, gateAt)
		assert.Contains(t, missing, "always-ok", "a pre-deploy probe run says nothing about this deploy")
	})

	t.Run("all probes reported in-window satisfies coverage", func(t *testing.T) {
		checks := []activities.ComponentHealthCheckRow{
			check("HTTPProbe", "public-endpoint", gateAt.Add(40*time.Second)),
			check("ExecProbe", "always-ok", gateAt.Add(40*time.Second)),
			check("ExecProbe", "gate-test-always-fails", gateAt.Add(40*time.Second)),
		}
		assert.Empty(t, missingProbes(declared, checks, gateAt))
	})

	t.Run("no declared probes means no coverage requirement", func(t *testing.T) {
		assert.Empty(t, missingProbes(nil, nil, gateAt))
	})
}

// Removed is recomputed each poll, so it covers both directions: deleting a
// probe labels its rows, and re-adding it clears the label next poll.
func TestMarkRemovedCheckRows(t *testing.T) {
	t.Parallel()

	rows := func() []activities.ComponentHealthCheckRow {
		return []activities.ComponentHealthCheckRow{
			{Kind: "ExecProbe", Name: "always-ok", Health: "healthy"},
			{Kind: "ExecProbe", Name: "gate-test-always-fails", Health: "unhealthy"},
			{Kind: "Deployment", Name: "whoami", Health: "healthy"},
			{Kind: "CustomCheck", Name: "checkout-latency", Health: "healthy"},
		}
	}

	removed := rows()
	markRemovedCheckRows(removed, []string{"always-ok"})
	assert.False(t, removed[0].Removed, "declared probe stays live")
	assert.True(t, removed[1].Removed, "undeclared probe is labelled removed")
	assert.False(t, removed[2].Removed, "k8s resources are never labelled")
	assert.False(t, removed[3].Removed, "custom checks are pushed, not declared")

	readded := rows()
	markRemovedCheckRows(readded, []string{"always-ok", "gate-test-always-fails"})
	assert.False(t, readded[1].Removed, "re-adding the probe clears the label immediately")
}

// A passing probe only proves an endpoint answered, not that the rollout
// succeeded — without cluster evidence the gate must not close on probes alone.
func TestJudgeWindowRequiresClusterEvidence(t *testing.T) {
	t.Parallel()

	gateAt := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	window := time.Minute
	probeOnly := []activities.GateHealthReport{
		{ObservedAtTS: gateAt.Add(50 * time.Second).Unix(), Health: "healthy", ClusterEvidence: false},
	}

	t.Run("probe-only cannot pass when cluster evidence is required", func(t *testing.T) {
		outcome, _ := judgeWindow(probeOnly, gateAt, gateAt.Add(window), window, true)
		assert.Equal(t, windowNeedData, outcome,
			"without cluster observations the gate has nothing to verify")
	})

	t.Run("probe-only passes when the component never had cluster evidence", func(t *testing.T) {
		outcome, _ := judgeWindow(probeOnly, gateAt, gateAt.Add(window), window, false)
		assert.Equal(t, windowPass, outcome,
			"a component owning no watched resources must stay verifiable by its probes")
	})

	t.Run("cluster evidence passes", func(t *testing.T) {
		reports := []activities.GateHealthReport{
			{ObservedAtTS: gateAt.Add(50 * time.Second).Unix(), Health: "healthy", ClusterEvidence: true},
		}
		outcome, _ := judgeWindow(reports, gateAt, gateAt.Add(window), window, true)
		assert.Equal(t, windowPass, outcome)
	})

	t.Run("a bad probe still fails fast regardless of source", func(t *testing.T) {
		reports := []activities.GateHealthReport{
			{ObservedAtTS: gateAt.Add(5 * time.Second).Unix(), Health: "unhealthy", ClusterEvidence: false},
		}
		outcome, report := judgeWindow(reports, gateAt, gateAt.Add(10*time.Second), window, true)
		assert.Equal(t, windowFailBad, outcome,
			"a failing check is a real signal whether or not it came from the cluster")
		assert.NotNil(t, report)
	})
}
