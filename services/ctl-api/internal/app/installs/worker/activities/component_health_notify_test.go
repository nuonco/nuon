package activities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestComponentHealthNotification(t *testing.T) {
	healthy := app.InstallComponentHealthStatusHealthy
	progressing := app.InstallComponentHealthStatusProgressing
	degraded := app.InstallComponentHealthStatusDegraded
	unhealthy := app.InstallComponentHealthStatusUnhealthy
	unknown := app.InstallComponentHealthStatusUnknown
	notApplicable := app.InstallComponentHealthStatusNotApplicable
	unset := app.InstallComponentHealthStatusUnset

	tests := []struct {
		name          string
		prior         app.InstallComponentHealthStatus
		verdict       app.InstallComponentHealthStatus
		wantNotify    bool
		wantRecovered bool
	}{
		{"healthy to degraded alerts", healthy, degraded, true, false},
		{"healthy to unhealthy alerts", healthy, unhealthy, true, false},
		{"unset to unhealthy alerts", unset, unhealthy, true, false},
		{"not-applicable to unhealthy alerts", notApplicable, unhealthy, true, false},
		{"unknown to unhealthy alerts", unknown, unhealthy, true, false},

		{"unhealthy to healthy resolves", unhealthy, healthy, true, true},
		{"degraded to healthy resolves", degraded, healthy, true, true},

		// Escalation and de-escalation inside the bad band are silent: the
		// subscriber already knows the component is broken.
		{"degraded to unhealthy is silent", degraded, unhealthy, false, false},
		{"unhealthy to degraded is silent", unhealthy, degraded, false, false},

		// Losing visibility is not a recovery and is not an alert.
		{"degraded to unknown is silent", degraded, unknown, false, false},
		{"unhealthy to unknown is silent", unhealthy, unknown, false, false},
		{"healthy to unknown is silent", healthy, unknown, false, false},
		{"unknown to healthy is silent", unknown, healthy, false, false},

		{"healthy to progressing is silent", healthy, progressing, false, false},
		{"progressing to healthy is silent", progressing, healthy, false, false},
		{"unhealthy to not-applicable is silent", unhealthy, notApplicable, false, false},
	}

	ic := &app.InstallComponent{
		ID:          "icmp1",
		ComponentID: "cmp1",
		Component:   app.Component{Name: "api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, ok := componentHealthNotification(ic, tt.prior, tt.verdict, "detail", nil)
			require.Equal(t, tt.wantNotify, ok)
			if !ok {
				return
			}
			assert.Equal(t, tt.wantRecovered, n.Recovered)
			assert.Equal(t, string(tt.verdict), n.Health)
			assert.Equal(t, string(tt.prior), n.PreviousHealth)
			assert.Equal(t, "api", n.ComponentName)
		})
	}
}

func TestInstallHealthNotification(t *testing.T) {
	healthy := app.InstallComponentHealthStatusHealthy
	degraded := app.InstallComponentHealthStatusDegraded
	unhealthy := app.InstallComponentHealthStatusUnhealthy
	unknown := app.InstallComponentHealthStatusUnknown
	notApplicable := app.InstallComponentHealthStatusNotApplicable

	t.Run("crossing into bad notifies with counts", func(t *testing.T) {
		n := installHealthNotification(
			[]app.InstallComponentHealthStatus{healthy, healthy, healthy},
			[]app.InstallComponentHealthStatus{healthy, degraded, unhealthy},
		)
		require.NotNil(t, n)
		assert.Equal(t, string(unhealthy), n.Health)
		assert.Equal(t, string(healthy), n.PreviousHealth)
		assert.Equal(t, 1, n.UnhealthyComponentCount)
		assert.Equal(t, 1, n.DegradedComponentCount)
	})

	t.Run("returning to healthy notifies", func(t *testing.T) {
		n := installHealthNotification(
			[]app.InstallComponentHealthStatus{healthy, unhealthy},
			[]app.InstallComponentHealthStatus{healthy, healthy},
		)
		require.NotNil(t, n)
		assert.Equal(t, string(healthy), n.Health)
		assert.Equal(t, string(unhealthy), n.PreviousHealth)
		assert.Zero(t, n.UnhealthyComponentCount)
	})

	t.Run("staying inside the bad band is silent", func(t *testing.T) {
		assert.Nil(t, installHealthNotification(
			[]app.InstallComponentHealthStatus{degraded},
			[]app.InstallComponentHealthStatus{unhealthy},
		))
	})

	t.Run("losing visibility is not a recovery", func(t *testing.T) {
		assert.Nil(t, installHealthNotification(
			[]app.InstallComponentHealthStatus{unhealthy},
			[]app.InstallComponentHealthStatus{unknown},
		))
	})

	t.Run("components with no health signal never notify", func(t *testing.T) {
		assert.Nil(t, installHealthNotification(
			[]app.InstallComponentHealthStatus{notApplicable},
			[]app.InstallComponentHealthStatus{notApplicable},
		))
	})
}

func TestDiagnosisFromDetails(t *testing.T) {
	tests := []struct {
		name    string
		details string
		want    string
	}{
		{"empty", "", ""},
		{"no diagnosis key", `{"status":{"phase":"Running"}}`, ""},
		{"unparseable", `not json`, ""},
		{
			name:    "extracts diagnosis only",
			details: `{"diagnosis":{"containers":[{"name":"api","last_termination_reason":"OOMKilled"}]},"status":{"phase":"Running"}}`,
			want:    `{"containers":[{"name":"api","last_termination_reason":"OOMKilled"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, diagnosisFromDetails(tt.details))
		})
	}
}
