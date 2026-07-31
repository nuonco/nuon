package awaitcomponenthealthy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

// A required check is one the runner cannot produce, so the gate has to hold
// until something external pushes it — that is the whole point of declaring it.
func TestRequiredCheckBlocksUntilPushed(t *testing.T) {
	gateStart := time.Now().Add(-2 * time.Minute)
	declared := []string{"migrations-applied", "public-endpoint"}

	t.Run("nothing pushed yet", func(t *testing.T) {
		assert.ElementsMatch(t, declared, missingProbes(declared, nil, gateStart))
	})

	t.Run("pushed before the gate opened does not count", func(t *testing.T) {
		stale := []activities.ComponentHealthCheckRow{
			{Name: "migrations-applied", ObservedAtTS: gateStart.Add(-time.Minute).Unix()},
		}
		assert.ElementsMatch(t, declared, missingProbes(declared, stale, gateStart),
			"a report from a previous deploy must not satisfy this one")
	})

	t.Run("one pushed, one still missing", func(t *testing.T) {
		rows := []activities.ComponentHealthCheckRow{
			{Name: "migrations-applied", ObservedAtTS: gateStart.Add(time.Minute).Unix()},
		}
		assert.Equal(t, []string{"public-endpoint"}, missingProbes(declared, rows, gateStart))
	})

	t.Run("all pushed since the gate opened", func(t *testing.T) {
		rows := []activities.ComponentHealthCheckRow{
			{Name: "migrations-applied", ObservedAtTS: gateStart.Add(time.Minute).Unix()},
			{Name: "public-endpoint", ObservedAtTS: gateStart.Add(time.Minute).Unix()},
		}
		assert.Empty(t, missingProbes(declared, rows, gateStart))
	})
}

// The snapshot must show what the gate is waiting on. Showing only resource
// verdicts made a blocked gate read as fully healthy.
func TestWithAwaitedChecksShowsUnknown(t *testing.T) {
	gateStart := time.Now().Add(-2 * time.Minute)
	declared := []string{"migrations-applied"}

	t.Run("never reported appears as unknown", func(t *testing.T) {
		rows := []activities.ComponentHealthCheckRow{
			{Kind: "Certificate", Name: "cert", Health: "healthy", ObservedAtTS: gateStart.Add(time.Minute).Unix()},
		}
		out := withAwaitedChecks(rows, declared, gateStart)
		require.Len(t, out, 2)
		assert.Equal(t, "healthy", out[0].Health, "an in-window resource report stands")
		assert.Equal(t, "migrations-applied", out[1].Name)
		assert.Equal(t, "unknown", out[1].Health)
	})

	// The window judges this deploy, so a resource verdict from before the apply
	// is not evidence about it — the same rule as for pushed checks.
	t.Run("a resource row from before the gate is unknown too", func(t *testing.T) {
		rows := []activities.ComponentHealthCheckRow{
			{Kind: "Certificate", Name: "cert", Health: "healthy", ObservedAtTS: gateStart.Add(-time.Hour).Unix()},
			{Kind: "Issuer", Name: "iss", Health: "healthy"},
		}
		out := withAwaitedChecks(rows, nil, gateStart)
		require.Len(t, out, 2)
		assert.Equal(t, "unknown", out[0].Health)
		assert.Equal(t, "unknown", out[1].Health, "a row with no timestamp is not in-window evidence")
	})

	t.Run("a pre-apply report is also unknown", func(t *testing.T) {
		rows := []activities.ComponentHealthCheckRow{
			{Kind: "CustomCheck", Name: "migrations-applied", Health: "healthy",
				ObservedAtTS: gateStart.Add(-time.Minute).Unix()},
		}
		out := withAwaitedChecks(rows, declared, gateStart)
		require.Len(t, out, 1, "the existing row is relabelled, not duplicated")
		assert.Equal(t, "unknown", out[0].Health)
		assert.Contains(t, out[0].Message, "since the deploy applied")
	})

	t.Run("reported since the gate opened is left alone", func(t *testing.T) {
		rows := []activities.ComponentHealthCheckRow{
			{Kind: "CustomCheck", Name: "migrations-applied", Health: "healthy",
				ObservedAtTS: gateStart.Add(time.Minute).Unix()},
		}
		out := withAwaitedChecks(rows, declared, gateStart)
		require.Len(t, out, 1)
		assert.Equal(t, "healthy", out[0].Health)
	})
}

// The countdown must not defeat the dedupe: two polls that differ only by the
// clock are the same state and should post one timeline entry, not one per tick.
func TestStripRemainingCollapsesCountdown(t *testing.T) {
	a := "healthy so far — 39s of the 1m0s window left — healthy (Certificate cert)"
	b := "healthy so far — 13s of the 1m0s window left — healthy (Certificate cert)"
	assert.Equal(t, stripRemaining(a), stripRemaining(b))

	// A real state change still reads as different.
	c := "waiting for the first health report since the apply — 50s of the window left"
	assert.NotEqual(t, stripRemaining(a), stripRemaining(c))

	// And a changed verdict is a different state even at the same remaining time.
	d := "healthy so far — 39s of the 1m0s window left — degraded (Certificate cert)"
	assert.NotEqual(t, stripRemaining(a), stripRemaining(d))
}

// The timeline should read as a record of what moved, not one line per poll.
func TestCheckTransitions(t *testing.T) {
	prev := map[string]string{}
	rows := func(h ...string) []activities.ComponentHealthCheckRow {
		return []activities.ComponentHealthCheckRow{
			{Name: "cert", Health: h[0]},
			{Name: "migrations-applied", Health: h[1]},
		}
	}

	// Everything starts unknown, so the first poll says nothing.
	first := rows("unknown", "unknown")
	assert.Empty(t, checkTransitions(prev, first))
	rememberCheckHealth(prev, first)

	// A quiet poll stays quiet.
	assert.Empty(t, checkTransitions(prev, rows("unknown", "unknown")))

	// Only what moved is reported, and it names both ends.
	moved := rows("healthy", "unknown")
	assert.Equal(t, "cert unknown → healthy", checkTransitions(prev, moved))
	rememberCheckHealth(prev, moved)

	both := rows("degraded", "healthy")
	assert.Equal(t, "cert healthy → degraded, migrations-applied unknown → healthy",
		checkTransitions(prev, both))
}

// A pass line naming one resource reads as though that was all it looked at.
func TestDescribeChecksSummarisesTheSet(t *testing.T) {
	checks := []activities.ComponentHealthCheckRow{
		{Kind: "Certificate", Name: "cert", Health: "healthy"},
		{Kind: "Issuer", Name: "iss", Health: "healthy"},
		{Kind: "ConfigMap", Name: "cm", Health: "not-applicable"},
		{Kind: "CustomCheck", Name: "migrations-applied", Health: "healthy"},
	}
	assert.Equal(t, "4 checks: 3 healthy, 1 not-applicable", describeChecks(checks))
	assert.Equal(t, "no checks reported", describeChecks(nil))
}

// The bug this pins: missingProbes only tests presence, so a required check that
// reported failing satisfied the gate and the deploy passed with an unhealthy
// check on screen. Reporting is not passing.
func TestFailedChecksFailTheWindow(t *testing.T) {
	gateStart := time.Now().Add(-2 * time.Minute)
	declared := []string{"migrations-applied", "smoke-tests"}
	inWindow := gateStart.Add(time.Minute).Unix()

	rows := []activities.ComponentHealthCheckRow{
		{Kind: "CustomCheck", Name: "migrations-applied", Health: "degraded", ObservedAtTS: inWindow},
		{Kind: "CustomCheck", Name: "smoke-tests", Health: "unhealthy", ObservedAtTS: inWindow},
	}

	// Both reported, so nothing is "missing" — which is exactly why presence
	// alone was not enough.
	assert.Empty(t, missingProbes(declared, rows, gateStart))
	assert.Equal(t,
		[]string{"migrations-applied is degraded", "smoke-tests is unhealthy"},
		failedChecks(declared, rows, gateStart))

	t.Run("healthy and not-applicable do not fail", func(t *testing.T) {
		ok := []activities.ComponentHealthCheckRow{
			{Name: "migrations-applied", Health: "healthy", ObservedAtTS: inWindow},
			{Name: "smoke-tests", Health: "not-applicable", ObservedAtTS: inWindow},
		}
		assert.Empty(t, failedChecks(declared, ok, gateStart))
	})

	t.Run("a failing report from before the gate is not this deploy's evidence", func(t *testing.T) {
		stale := []activities.ComponentHealthCheckRow{
			{Name: "smoke-tests", Health: "unhealthy", ObservedAtTS: gateStart.Add(-time.Hour).Unix()},
		}
		assert.Empty(t, failedChecks(declared, stale, gateStart))
	})

	t.Run("undeclared failing checks are not the gate's business", func(t *testing.T) {
		other := []activities.ComponentHealthCheckRow{
			{Name: "some-other-check", Health: "unhealthy", ObservedAtTS: inWindow},
		}
		assert.Empty(t, failedChecks(declared, other, gateStart))
	})
}
