package app

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func healthHstore(statuses ...InstallComponentHealthStatus) pgtype.Hstore {
	out := pgtype.Hstore{}
	for i, s := range statuses {
		v := string(s)
		out[string(rune('a'+i))] = &v
	}
	return out
}

func TestCompositeComponentHealthStatus(t *testing.T) {
	tests := []struct {
		name     string
		statuses pgtype.Hstore
		want     InstallComponentHealthStatus
	}{
		{"no components", pgtype.Hstore{}, InstallComponentHealthStatusUnset},
		{"only unset", healthHstore(InstallComponentHealthStatusUnset), InstallComponentHealthStatusUnset},
		{"only not-applicable", healthHstore(InstallComponentHealthStatusNotApplicable), InstallComponentHealthStatusUnset},
		{"all healthy", healthHstore(InstallComponentHealthStatusHealthy, InstallComponentHealthStatusHealthy), InstallComponentHealthStatusHealthy},
		{"unhealthy wins", healthHstore(InstallComponentHealthStatusHealthy, InstallComponentHealthStatusDegraded, InstallComponentHealthStatusUnhealthy), InstallComponentHealthStatusUnhealthy},
		{"degraded over unknown", healthHstore(InstallComponentHealthStatusUnknown, InstallComponentHealthStatusDegraded), InstallComponentHealthStatusDegraded},
		{"unknown over progressing", healthHstore(InstallComponentHealthStatusProgressing, InstallComponentHealthStatusUnknown), InstallComponentHealthStatusUnknown},
		{"progressing over healthy", healthHstore(InstallComponentHealthStatusHealthy, InstallComponentHealthStatusProgressing), InstallComponentHealthStatusProgressing},
		{"not-applicable excluded", healthHstore(InstallComponentHealthStatusNotApplicable, InstallComponentHealthStatusHealthy), InstallComponentHealthStatusHealthy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := compositeComponentHealthStatus(tt.statuses)
			assert.Equal(t, tt.want, got)
		})
	}
}
