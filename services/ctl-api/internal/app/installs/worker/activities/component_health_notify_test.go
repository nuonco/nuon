package activities

import (
	"testing"
	"time"

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
		priorAlerted  bool
		wantNotify    bool
		wantRecovered bool
	}{
		{"healthy to degraded alerts", healthy, degraded, false, true, false},
		{"healthy to unhealthy alerts", healthy, unhealthy, false, true, false},
		{"unset to unhealthy alerts", unset, unhealthy, false, true, false},
		{"not-applicable to unhealthy alerts", notApplicable, unhealthy, false, true, false},
		{"unknown to unhealthy alerts", unknown, unhealthy, false, true, false},

		{"unhealthy to healthy resolves", unhealthy, healthy, true, true, true},
		{"degraded to healthy resolves", degraded, healthy, true, true, true},

		// Escalation and de-escalation inside the bad band are silent: the
		// subscriber already knows the component is broken.
		{"degraded to unhealthy is silent", degraded, unhealthy, true, false, false},
		{"unhealthy to degraded is silent", unhealthy, degraded, true, false, false},

		// Alerts/resolutions pair via the alerted flag: a suppressed spell
		// resolves silently, a late one alerts once lifted, a steady spell is quiet.
		{"suppressed spell resolves silently", unhealthy, healthy, false, false, false},
		{"suppression lifted mid-spell alerts late", unhealthy, unhealthy, false, true, false},
		{"steady alerted spell is silent", unhealthy, unhealthy, true, false, false},

		// Losing visibility is not a recovery and is not an alert.
		{"degraded to unknown is silent", degraded, unknown, true, false, false},
		{"unhealthy to unknown is silent", unhealthy, unknown, true, false, false},
		{"healthy to unknown is silent", healthy, unknown, false, false, false},
		{"unknown to healthy is silent", unknown, healthy, false, false, false},

		{"healthy to progressing is silent", healthy, progressing, false, false, false},
		{"progressing to healthy is silent", progressing, healthy, false, false, false},
		{"unhealthy to not-applicable is silent", unhealthy, notApplicable, true, false, false},
	}

	ic := &app.InstallComponent{
		ID:          "icmp1",
		ComponentID: "cmp1",
		Component:   app.Component{Name: "api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, ok := componentHealthNotification(ic, tt.prior, tt.verdict, tt.priorAlerted, "detail", nil)
			require.Equal(t, tt.wantNotify, ok)
			if !ok {
				return
			}
			assert.Equal(t, tt.wantRecovered, n.Recovered)
			assert.Equal(t, string(tt.verdict), n.Health)
			assert.Equal(t, "api", n.ComponentName)
			if tt.prior != tt.verdict {
				assert.Equal(t, string(tt.prior), n.PreviousHealth)
			} else {
				assert.Empty(t, n.PreviousHealth,
					"a late alert has no transition to report")
			}
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

// markDownstream's graph logic is what turns "postgres down plus six
// crashlooping dependents" into one alert instead of seven.
func TestMarkDownstreamLabelling(t *testing.T) {
	unhealthy := app.InstallComponentHealthStatusUnhealthy
	degraded := app.InstallComponentHealthStatusDegraded
	healthy := app.InstallComponentHealthStatusHealthy

	eval := func(componentID, name string, verdict app.InstallComponentHealthStatus) componentEval {
		return componentEval{
			ic: &app.InstallComponent{
				ID:          "ic-" + componentID,
				ComponentID: componentID,
				Component:   app.Component{Name: name},
			},
			verdict: verdict,
		}
	}

	tests := []struct {
		name string
		// deps maps componentID -> dependency componentIDs
		deps  map[string][]string
		evals []componentEval
		// want maps component name -> expected downstreamOf
		want map[string]string
	}{
		{
			name:  "a lone failure is its own root cause",
			deps:  map[string][]string{"api": {"pg"}},
			evals: []componentEval{eval("pg", "postgres", healthy), eval("api", "api", unhealthy)},
			want:  map[string]string{"api": ""},
		},
		{
			name:  "dependent of a bad dependency is labelled downstream",
			deps:  map[string][]string{"api": {"pg"}},
			evals: []componentEval{eval("pg", "postgres", unhealthy), eval("api", "api", unhealthy)},
			want:  map[string]string{"postgres": "", "api": "postgres"},
		},
		{
			name: "one root cause labels every dependent",
			deps: map[string][]string{"api": {"pg"}, "web": {"pg"}, "worker": {"pg"}},
			evals: []componentEval{
				eval("pg", "postgres", unhealthy),
				eval("api", "api", unhealthy),
				eval("web", "web", degraded),
				eval("worker", "worker", unhealthy),
			},
			want: map[string]string{"postgres": "", "api": "postgres", "web": "postgres", "worker": "postgres"},
		},
		{
			name:  "a healthy dependency does not absorb the blame",
			deps:  map[string][]string{"api": {"cache"}},
			evals: []componentEval{eval("cache", "cache", healthy), eval("api", "api", unhealthy), eval("web", "web", degraded)},
			want:  map[string]string{"api": "", "web": ""},
		},
		{
			name:  "chain blames the nearest bad dependency",
			deps:  map[string][]string{"api": {"pg"}, "web": {"api"}},
			evals: []componentEval{eval("pg", "postgres", unhealthy), eval("api", "api", unhealthy), eval("web", "web", unhealthy)},
			want:  map[string]string{"postgres": "", "api": "postgres", "web": "api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evals := tt.evals
			markDownstreamWithDeps(evals, tt.deps)

			got := map[string]string{}
			for i := range evals {
				got[evals[i].ic.Component.Name] = evals[i].downstreamOf
			}
			for name, want := range tt.want {
				assert.Equal(t, want, got[name], "component %s", name)
			}
		})
	}
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

// Regression: an undeployed component must never get a health verdict. Found
// on stage — probes reported `whoami` healthy before it had ever been deployed.
func TestComponentVerdictRequiresADeploy(t *testing.T) {
	now := time.Now()
	healthyReports := []componentHealthReport{{
		ObservedAt:     now,
		Health:         app.InstallComponentHealthStatusHealthy,
		Resources:      3,
		ResourceCounts: map[string]int{"healthy": 3},
	}}

	tests := []struct {
		name   string
		status app.InstallComponentStatus
		want   app.InstallComponentHealthStatus
	}{
		{"never deployed (unset)", app.InstallComponentStatusUnset, app.InstallComponentHealthStatusNotApplicable},
		{"queued", app.InstallComponentStatusQueued, app.InstallComponentHealthStatusNotApplicable},
		{"pending", app.InstallComponentStatusPending, app.InstallComponentHealthStatusNotApplicable},
		{"planning", app.InstallComponentStatusPlanning, app.InstallComponentHealthStatusNotApplicable},
		{"executing", app.InstallComponentStatusExecuting, app.InstallComponentHealthStatusNotApplicable},
		{"disabled", app.InstallComponentStatusDisabled, app.InstallComponentHealthStatusNotApplicable},
		{"torn down", app.InstallComponentStatusInactive, app.InstallComponentHealthStatusNotApplicable},
		{"deleted", app.InstallComponentStatusDeleted, app.InstallComponentHealthStatusNotApplicable},

		// Deployed at least once: health applies.
		{"active", app.InstallComponentStatusActive, app.InstallComponentHealthStatusHealthy},
		{"noop deploy", app.InstallComponentStatusNoop, app.InstallComponentHealthStatusHealthy},
		{"failed deploy still reports health", app.InstallComponentStatusError, app.InstallComponentHealthStatusHealthy},
	}

	a := &Activities{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ic := &app.InstallComponent{
				ID:          "icmp1",
				ComponentID: "cmp1",
				Status:      tt.status,
				Component:   app.Component{Name: "whoami", Type: app.ComponentTypeHelmChart},
			}
			assert.Equal(t, tt.want, a.componentVerdict(ic, healthyReports, now))
		})
	}
}

func TestHasDeployed(t *testing.T) {
	assert.True(t, app.InstallComponentStatusActive.HasDeployed())
	assert.True(t, app.InstallComponentStatusNoop.HasDeployed())
	assert.True(t, app.InstallComponentStatusError.HasDeployed())

	for _, s := range []app.InstallComponentStatus{
		app.InstallComponentStatusUnset,
		app.InstallComponentStatusQueued,
		app.InstallComponentStatusPending,
		app.InstallComponentStatusPlanning,
		app.InstallComponentStatusSyncing,
		app.InstallComponentStatusExecuting,
		app.InstallComponentStatusDisabled,
		app.InstallComponentStatusInactive,
		app.InstallComponentStatusDeleted,
		app.InstallComponentStatusUnknown,
	} {
		assert.False(t, s.HasDeployed(), "status %q must not count as deployed", s)
	}
}
