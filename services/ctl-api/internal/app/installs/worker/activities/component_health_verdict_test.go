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
		{"unknown recovers immediately on fresh report", unknown, healthReports(now, degraded), degraded},
		{"not-applicable flips immediately once observed", notApplicable, healthReports(now, progressing), progressing},

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
