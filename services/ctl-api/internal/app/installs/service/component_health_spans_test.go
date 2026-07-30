package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func transitionAt(observedAt time.Time, toHealth string) app.InstallComponentHealthTransition {
	return app.InstallComponentHealthTransition{ObservedAt: observedAt, ToHealth: toHealth}
}

func TestHealthSpans(t *testing.T) {
	day0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("no transitions leaves the whole window unknown", func(t *testing.T) {
		from, to := day0, day0.Add(2*time.Hour)

		spans := healthSpans(nil, from, to, healthUnknown)

		assert.Equal(t, []healthSpan{{From: from, To: to, Health: healthUnknown}}, spans)
	})

	t.Run("one transition mid-window splits into unknown then the new verdict", func(t *testing.T) {
		from, to := day0, day0.Add(2*time.Hour)
		mid := day0.Add(1 * time.Hour)

		spans := healthSpans([]app.InstallComponentHealthTransition{
			transitionAt(mid, "healthy"),
		}, from, to, healthUnknown)

		assert.Equal(t, []healthSpan{
			{From: from, To: mid, Health: healthUnknown},
			{From: mid, To: to, Health: "healthy"},
		}, spans)
	})

	t.Run("a bad span still open at now is clipped to the window end", func(t *testing.T) {
		from, to := day0, day0.Add(3*time.Hour)
		wentBad := day0.Add(2 * time.Hour)

		spans := healthSpans([]app.InstallComponentHealthTransition{
			transitionAt(wentBad, "unhealthy"),
		}, from, to, healthUnknown)

		assert.Equal(t, []healthSpan{
			{From: from, To: wentBad, Health: healthUnknown},
			{From: wentBad, To: to, Health: "unhealthy"},
		}, spans)
		assert.True(t, spans[len(spans)-1].To.Equal(to), "trailing span must be clipped exactly at `to`, not extend past it")
	})

	t.Run("a window entirely unknown collapses to a single span via carry-forward, not the pre-transition gap", func(t *testing.T) {
		from, to := day0, day0.Add(4*time.Hour)

		spans := healthSpans([]app.InstallComponentHealthTransition{
			transitionAt(from, "unknown"),
		}, from, to, healthUnknown)

		assert.Equal(t, []healthSpan{{From: from, To: to, Health: "unknown"}}, spans)
	})

	t.Run("transitions outside [from, to) are ignored", func(t *testing.T) {
		from, to := day0.Add(1*time.Hour), day0.Add(2*time.Hour)

		spans := healthSpans([]app.InstallComponentHealthTransition{
			transitionAt(day0, "unhealthy"),                  // before `from`
			transitionAt(day0.Add(3*time.Hour), "unhealthy"), // at/after `to`
		}, from, to, healthUnknown)

		assert.Equal(t, []healthSpan{{From: from, To: to, Health: healthUnknown}}, spans)
	})

	t.Run("an empty or inverted window yields no spans", func(t *testing.T) {
		assert.Nil(t, healthSpans(nil, day0, day0, healthUnknown))
		assert.Nil(t, healthSpans(nil, day0.Add(time.Hour), day0, healthUnknown))
	})
}

func TestFoldDailyHealthCrossesMidnight(t *testing.T) {
	day0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	from := day0
	to := day0.AddDate(0, 0, 2) // two full calendar days

	transitions := []app.InstallComponentHealthTransition{
		transitionAt(day0, "healthy"),
		transitionAt(day0.Add(23*time.Hour), "degraded"),                // 23:00 day0
		transitionAt(day0.AddDate(0, 0, 1).Add(1*time.Hour), "healthy"), // 01:00 day1, 2h degraded span
	}

	spans := healthSpans(transitions, from, to, healthUnknown)
	daily := foldDailyHealth(spans, from, 2)

	assert.Len(t, daily, 2)

	assert.Equal(t, "2026-01-01", daily[0].Date)
	assert.Equal(t, "degraded", daily[0].Health)
	assert.Equal(t, int64(3600), daily[0].DegradedSeconds)
	assert.Equal(t, int64(0), daily[0].UnknownSeconds)
	assert.Equal(t, int64(86400), daily[0].ObservedSeconds)

	assert.Equal(t, "2026-01-02", daily[1].Date)
	assert.Equal(t, "degraded", daily[1].Health)
	assert.Equal(t, int64(3600), daily[1].DegradedSeconds)
	assert.Equal(t, int64(0), daily[1].UnknownSeconds)
	assert.Equal(t, int64(86400), daily[1].ObservedSeconds)
}

func TestFoldDailyHealthBackToBackSameDayFlips(t *testing.T) {
	day0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	from := day0
	to := day0.AddDate(0, 0, 1)

	transitions := []app.InstallComponentHealthTransition{
		transitionAt(day0, "healthy"),
		transitionAt(day0.Add(10*time.Hour), "degraded"),
		transitionAt(day0.Add(10*time.Hour+5*time.Minute), "unhealthy"),
		transitionAt(day0.Add(10*time.Hour+10*time.Minute), "healthy"),
	}

	spans := healthSpans(transitions, from, to, healthUnknown)
	daily := foldDailyHealth(spans, from, 1)

	assert.Len(t, daily, 1)
	assert.Equal(t, "2026-01-01", daily[0].Date)
	// unhealthy briefly beat out degraded and healthy despite lasting 5 minutes.
	assert.Equal(t, "unhealthy", daily[0].Health)
	assert.Equal(t, int64(300), daily[0].DegradedSeconds)
	assert.Equal(t, int64(300), daily[0].UnhealthySeconds)
	assert.Equal(t, int64(0), daily[0].UnknownSeconds)
	assert.Equal(t, int64(86400), daily[0].ObservedSeconds)
}

func TestFoldDailyHealthNoDataLeavesHealthEmpty(t *testing.T) {
	day0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	daily := foldDailyHealth(nil, day0, 3)

	assert.Len(t, daily, 3)
	for i, d := range daily {
		assert.Equal(t, day0.AddDate(0, 0, i).Format("2006-01-02"), d.Date)
		assert.Equal(t, "", d.Health)
		assert.Equal(t, int64(0), d.ObservedSeconds)
	}
}

func TestHealthTotalsUptimePercent(t *testing.T) {
	tests := []struct {
		name   string
		totals healthTotals
		want   float64
	}{
		{"window entirely unknown reports 0 uptime and is distinguishable via observedSeconds", healthTotals{UnknownSeconds: 3600}, 0},
		{"all healthy is 100%", healthTotals{HealthySeconds: 3600}, 100},
		{"half degraded is 50%", healthTotals{HealthySeconds: 1800, DegradedSeconds: 1800}, 50},
		{"progressing counts as up", healthTotals{ProgressingSeconds: 1800, HealthySeconds: 1800}, 100},
		{"unknown excluded from both numerator and denominator", healthTotals{HealthySeconds: 1800, UnhealthySeconds: 1800, UnknownSeconds: 100000}, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.totals.uptimePercent())
		})
	}

	t.Run("window entirely unknown also reports 0 observed seconds", func(t *testing.T) {
		totals := healthTotals{UnknownSeconds: 3600}
		assert.Equal(t, int64(0), totals.observedSeconds())
		assert.Equal(t, float64(0), totals.uptimePercent())
	})
}

func TestWorstDailyAcrossComponents(t *testing.T) {
	day0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	healthyComponent := []dailyHealthBucket{{Date: "2026-01-01", Health: "healthy", ObservedSeconds: 86400}}
	unhealthyComponent := []dailyHealthBucket{{Date: "2026-01-01", Health: "unhealthy", UnhealthySeconds: 86400, ObservedSeconds: 86400}}
	noDataComponent := []dailyHealthBucket{{Date: "2026-01-01"}}

	out := worstDailyAcrossComponents([][]dailyHealthBucket{healthyComponent, unhealthyComponent, noDataComponent}, day0, 1)

	assert.Len(t, out, 1)
	assert.Equal(t, "unhealthy", out[0].Health)
	assert.Equal(t, int64(86400), out[0].UnhealthySeconds)
}

// Pins a live failure: resetting the baseline on a stable healthy component left
// zero transitions in the window, so the whole window read as unknown.
func TestHealthSpansSeededFromPriorState(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 7, 30, 7, 25, 38, 0, time.UTC)
	to := from.Add(6 * time.Minute)

	spans := healthSpans(nil, from, to, "healthy")
	require.Len(t, spans, 1)
	assert.Equal(t, "healthy", spans[0].Health,
		"the verdict in effect before the window carries in — a minute of uptime is a minute of uptime")
	assert.Equal(t, from, spans[0].From)
	assert.Equal(t, to, spans[0].To)

	empty := healthSpans(nil, from, to, "")
	require.Len(t, empty, 1)
	assert.Equal(t, healthUnknown, empty[0].Health, "no seed means nothing can be claimed")
}
