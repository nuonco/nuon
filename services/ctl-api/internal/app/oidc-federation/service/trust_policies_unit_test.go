package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestParseTrustPolicyRole(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    app.RoleType
		wantErr bool
	}{
		{name: "empty defaults to read-only", want: app.RoleTypeOrgReadOnly},
		{name: "admin", raw: "org_admin", want: app.RoleTypeOrgAdmin},
		{name: "deprecated builder is rejected", raw: "org_builder", wantErr: true},
		{name: "runner rejected", raw: "runner", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTrustPolicyRole(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
