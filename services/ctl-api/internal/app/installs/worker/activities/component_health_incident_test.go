package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// replay walks reports oldest-first the way the evaluator sees them, newest at
// index 0, and returns the verdict after each tick.
func replay(start app.InstallComponentHealthStatus, now time.Time, oldestFirst []app.InstallComponentHealthStatus) []app.InstallComponentHealthStatus {
	verdict := start
	out := make([]app.InstallComponentHealthStatus, 0, len(oldestFirst))
	for i := range oldestFirst {
		window := make([]componentHealthReport, 0, i+1)
		for j := i; j >= 0; j-- {
			window = append(window, componentHealthReport{
				ObservedAt: now.Add(time.Duration(j-i) * time.Minute),
				Health:     oldestFirst[j],
			})
		}
		verdict = nextComponentHealthVerdict(verdict, window, now)
		out = append(out, verdict)
	}
	return out
}

// The incident: a deploy at 18:40:12 produced one degraded report at 18:41:23
// from an HPA waiting on metrics, and the component stayed degraded until
// 18:58:24 — 15m of event window plus two reports, for a fault that was over in
// about a minute. One transient report must now cost nothing.
func TestOneTransientReportDoesNotDegrade(t *testing.T) {
	now := time.Now()
	degraded := app.InstallComponentHealthStatusDegraded
	healthy := app.InstallComponentHealthStatusHealthy
	notApplicable := app.InstallComponentHealthStatusNotApplicable

	got := replay(notApplicable, now, []app.InstallComponentHealthStatus{
		degraded, healthy, healthy, healthy, healthy,
	})

	for i, v := range got {
		assert.NotEqual(t, degraded, v, "tick %d: a single transient must never reach a verdict", i)
	}
	assert.Equal(t, healthy, got[len(got)-1])
}

// A fault that persists is still reported, and recovery costs two reports
// rather than the fifteen minutes the event window used to add.
func TestSustainedFaultDegradesThenRecoversPromptly(t *testing.T) {
	now := time.Now()
	degraded := app.InstallComponentHealthStatusDegraded
	healthy := app.InstallComponentHealthStatusHealthy

	got := replay(healthy, now, []app.InstallComponentHealthStatus{
		degraded, degraded, degraded, degraded, healthy, healthy, healthy,
	})

	assert.Equal(t, healthy, got[0], "one bad report holds")
	assert.Equal(t, healthy, got[1], "two bad reports hold")
	assert.Equal(t, degraded, got[2], "three consecutive bad reports are earned")
	assert.Equal(t, degraded, got[3])
	assert.Equal(t, degraded, got[4], "one good report is not recovery")
	assert.Equal(t, healthy, got[5], "two good reports recover")
	assert.Equal(t, healthy, got[6])
}
