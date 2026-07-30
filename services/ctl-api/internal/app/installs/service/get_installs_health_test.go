package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func hstoreOf(kv map[string]string) pgtype.Hstore {
	h := make(pgtype.Hstore, len(kv))
	for k, v := range kv {
		v := v
		h[k] = &v
	}
	return h
}

func installWithHealth(health app.InstallComponentHealthStatus, componentStatuses map[string]string) app.Install {
	return app.Install{
		CompositeHealthStatus:   health,
		ComponentHealthStatuses: hstoreOf(componentStatuses),
	}
}

func TestSummarizeInstallsHealthCounts(t *testing.T) {
	installs := []app.Install{
		installWithHealth(app.InstallComponentHealthStatusHealthy, map[string]string{"c1": "healthy"}),
		installWithHealth(app.InstallComponentHealthStatusHealthy, map[string]string{"c1": "healthy"}),
		installWithHealth(app.InstallComponentHealthStatusDegraded, map[string]string{"c1": "degraded", "c2": "healthy"}),
		installWithHealth(app.InstallComponentHealthStatusUnhealthy, map[string]string{"c1": "unhealthy", "c2": "unhealthy"}),
		installWithHealth(app.InstallComponentHealthStatusUnknown, map[string]string{"c1": "unknown"}),
		installWithHealth(app.InstallComponentHealthStatusProgressing, map[string]string{"c1": "progressing"}),
		installWithHealth(app.InstallComponentHealthStatusUnset, nil),
		installWithHealth(app.InstallComponentHealthStatusNotApplicable, nil),
	}

	resp := summarizeInstallsHealth(installs)

	assert.Equal(t, 8, resp.Total)
	assert.Equal(t, 2, resp.Healthy)
	assert.Equal(t, 1, resp.Degraded)
	assert.Equal(t, 1, resp.Unhealthy)
	assert.Equal(t, 2, resp.Unknown, "progressing and unknown both roll up into the unknown bucket")
	assert.Equal(t, 2, resp.Unset, "unset and not-applicable both roll up into the unset bucket")
	assert.Len(t, resp.Installs, 8)
}

func TestSummarizeInstallsHealthPerComponentCounts(t *testing.T) {
	installs := []app.Install{
		installWithHealth(app.InstallComponentHealthStatusUnhealthy, map[string]string{
			"c1": "unhealthy",
			"c2": "degraded",
			"c3": "healthy",
		}),
	}

	resp := summarizeInstallsHealth(installs)

	got := resp.Installs[0]
	assert.Equal(t, 1, got.UnhealthyComponents)
	assert.Equal(t, 1, got.DegradedComponents)
}

func TestSummarizeInstallsHealthAllHealthy(t *testing.T) {
	tests := []struct {
		name     string
		installs []app.Install
		want     bool
	}{
		{
			name:     "no installs is never all_healthy",
			installs: nil,
			want:     false,
		},
		{
			name: "every install healthy is all_healthy",
			installs: []app.Install{
				installWithHealth(app.InstallComponentHealthStatusHealthy, nil),
				installWithHealth(app.InstallComponentHealthStatusHealthy, nil),
			},
			want: true,
		},
		{
			name: "a single degraded install breaks all_healthy",
			installs: []app.Install{
				installWithHealth(app.InstallComponentHealthStatusHealthy, nil),
				installWithHealth(app.InstallComponentHealthStatusDegraded, nil),
			},
			want: false,
		},
		{
			name: "an unset install must never let all_healthy report true, even alongside all-healthy peers",
			installs: []app.Install{
				installWithHealth(app.InstallComponentHealthStatusHealthy, nil),
				installWithHealth(app.InstallComponentHealthStatusHealthy, nil),
				installWithHealth(app.InstallComponentHealthStatusUnset, nil),
			},
			want: false,
		},
		{
			name: "all installs unset (no data anywhere) is not all_healthy",
			installs: []app.Install{
				installWithHealth(app.InstallComponentHealthStatusUnset, nil),
				installWithHealth(app.InstallComponentHealthStatusUnset, nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := summarizeInstallsHealth(tt.installs)
			assert.Equal(t, tt.want, resp.AllHealthy)
		})
	}
}
