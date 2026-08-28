package app

import (
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestInstallIndexesDoNotCoupleCustomerManagedState(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	for _, index := range (&Install{}).Indexes(db) {
		require.NotContains(t, index.Name, "customer_managed")
		require.NotContains(t, index.Columns, "customer_managed_registration_id")
	}
}

func TestInstallAfterQuerySelectsLoadedManagementRecords(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	install := Install{
		SandboxMode: sql.NullBool{Valid: true},
		ManagementPolicyVersions: []InstallManagementPolicyVersion{
			{ID: "new-policy"},
			{ID: "old-policy"},
		},
		InstallRegistrations: []InstallRegistration{
			{ID: "new-registration"},
			{ID: "old-registration"},
		},
	}

	require.NoError(t, install.AfterQuery(db))
	require.NotNil(t, install.ManagementPolicy)
	require.NotNil(t, install.LatestRegistration)
	assert.Equal(t, "new-policy", install.ManagementPolicy.ID)
	assert.Equal(t, "new-registration", install.LatestRegistration.ID)
}
