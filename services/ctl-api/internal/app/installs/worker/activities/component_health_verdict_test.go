package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func healthReports(now time.Time, healths ...app.InstallComponentHealthStatus) []componentHealthReport {
	reports := make([]componentHealthReport, 0, len(healths))
	for i, h := range healths {
		reports = append(reports, componentHealthReport{
			ObservedAt: now.Add(-time.Duration(i) * time.Minute),
			Health:     h,
		})
	}
	return reports
}

func TestNextComponentHealthVerdict(t *testing.T) {
	now := time.Now()

	healthy := app.InstallComponentHealthStatusHealthy
	progressing := app.InstallComponentHealthStatusProgressing
	degraded := app.InstallComponentHealthStatusDegraded
	unhealthy := app.InstallComponentHealthStatusUnhealthy
	unknown := app.InstallComponentHealthStatusUnknown
	notApplicable := app.InstallComponentHealthStatusNotApplicable
	unset := app.InstallComponentHealthStatusUnset

	tests := []struct {
		name    string
		current app.InstallComponentHealthStatus
		reports []componentHealthReport
		want    app.InstallComponentHealthStatus
	}{
		{"never observed stays not-applicable", unset, nil, notApplicable},
		{"not-applicable with no reports holds", notApplicable, nil, notApplicable},
		{"observed before but no reports means unknown", healthy, nil, unknown},
		{"stale reports mean unknown", healthy, healthReports(now.Add(-6*time.Minute), unhealthy), unknown},
		{"a single delayed report does not trip unknown", healthy, healthReports(now.Add(-4*time.Minute), healthy), healthy},
		{"first fresh report sets baseline immediately", unset, healthReports(now, healthy), healthy},
		{"not-applicable flips immediately once observed", notApplicable, healthReports(now, progressing), progressing},
		{"unknown adopts a good report immediately", unknown, healthReports(now, healthy), healthy},

		// A component fresh from a deploy has no baseline. Adopting one bad
		// report made every transient the runner caught first an outage.
		{"bootstrap does not claim bad on one report", notApplicable, healthReports(now, degraded), notApplicable},
		{"bootstrap does not claim bad on two reports", notApplicable, healthReports(now, degraded, degraded), notApplicable},
		{"bootstrap claims bad once earned", notApplicable, healthReports(now, degraded, degraded, degraded), degraded},
		{"unset does not claim bad on one report", unset, healthReports(now, unhealthy), unset},
		{"unknown does not claim bad on one report", unknown, healthReports(now, degraded), unknown},

		{"one bad report holds healthy", healthy, healthReports(now, unhealthy, healthy, healthy), healthy},
		{"two bad reports still hold healthy", healthy, healthReports(now, unhealthy, unhealthy, healthy), healthy},
		{"three consecutive bad reports flip unhealthy", healthy, healthReports(now, unhealthy, unhealthy, unhealthy), unhealthy},
		{"three bad reports of mixed severity flip to latest", healthy, healthReports(now, unhealthy, degraded, degraded), unhealthy},
		{"fewer reports than required hold", healthy, healthReports(now, unhealthy, unhealthy), healthy},

		{"one good report holds unhealthy", unhealthy, healthReports(now, healthy, unhealthy, unhealthy), unhealthy},
		{"two consecutive good reports recover", unhealthy, healthReports(now, healthy, healthy, unhealthy), healthy},
		{"recovery interrupted by progressing holds", unhealthy, healthReports(now, healthy, progressing, unhealthy), unhealthy},
		{"improvement to degraded needs two agreeing reports", unhealthy, healthReports(now, degraded, degraded, unhealthy), degraded},

		{"healthy to progressing needs two reports", healthy, healthReports(now, progressing, healthy), healthy},
		{"two progressing reports flip", healthy, healthReports(now, progressing, progressing), progressing},

		{"same verdict holds", degraded, healthReports(now, degraded, degraded), degraded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, nextComponentHealthVerdict(tt.current, tt.reports, now))
		})
	}
}

// Pins the gap a live install exposed: an ingress with no class never gets a
// load balancer address, so it reports progressing forever — and because
// progressing never alerts, 15 hours passed with nobody told.
func TestEscalateStuckProgressing(t *testing.T) {
	now := time.Now()
	progressingSince := func(d time.Duration) *app.InstallComponent {
		return &app.InstallComponent{
			HealthStatus: app.InstallComponentHealthStatusProgressing,
			HealthStatusV2: app.CompositeStatus{
				Status:      app.Status(app.InstallComponentHealthStatusProgressing),
				CreatedAtTS: now.Add(-d).Unix(),
			},
		}
	}

	t.Run("a slow rollout is left alone", func(t *testing.T) {
		got := escalateStuckProgressing(app.InstallComponentHealthStatusProgressing,
			progressingSince(10*time.Minute), now)
		assert.Equal(t, app.InstallComponentHealthStatusProgressing, got)
	})

	t.Run("stuck past the limit becomes degraded", func(t *testing.T) {
		got := escalateStuckProgressing(app.InstallComponentHealthStatusProgressing,
			progressingSince(16*time.Hour), now)
		assert.Equal(t, app.InstallComponentHealthStatusDegraded, got)
	})

	t.Run("a freshly progressing component has no elapsed time to judge", func(t *testing.T) {
		fresh := &app.InstallComponent{
			HealthStatus:   app.InstallComponentHealthStatusHealthy,
			HealthStatusV2: app.CompositeStatus{CreatedAtTS: now.Add(-16 * time.Hour).Unix()},
		}
		got := escalateStuckProgressing(app.InstallComponentHealthStatusProgressing, fresh, now)
		assert.Equal(t, app.InstallComponentHealthStatusProgressing, got,
			"the old timestamp belongs to the previous verdict, not this one")
	})

	t.Run("other verdicts pass through untouched", func(t *testing.T) {
		for _, v := range []app.InstallComponentHealthStatus{
			app.InstallComponentHealthStatusHealthy,
			app.InstallComponentHealthStatusUnknown,
			app.InstallComponentHealthStatusUnhealthy,
			app.InstallComponentHealthStatusNotApplicable,
		} {
			assert.Equal(t, v, escalateStuckProgressing(v, progressingSince(16*time.Hour), now))
		}
	})
}
