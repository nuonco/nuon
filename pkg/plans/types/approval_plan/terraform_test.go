package approvalplan

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTerraformApprovalPlan_IsNoop(t *testing.T) {
	tests := map[string]struct {
		planJSON string
		want     bool
	}{
		"empty plan json": {
			planJSON: ``,
			want:     false,
		},
		"no changes": {
			planJSON: `{"format_version": "1.2"}`,
			want:     true,
		},
		"only no-op changes": {
			planJSON: `{"format_version": "1.2", "resource_changes": [
				{"address": "aws_rds_cluster.main", "mode": "managed", "change": {"actions": ["no-op"]}}
			]}`,
			want: true,
		},
		"delete": {
			planJSON: `{"format_version": "1.2", "resource_changes": [
				{"address": "aws_rds_cluster.main", "mode": "managed", "change": {"actions": ["delete"]}}
			]}`,
			want: false,
		},
		"teardown that only retains resources": {
			planJSON: `{"format_version": "1.2", "resource_changes": [
				{"address": "aws_rds_cluster.main", "mode": "managed", "change": {"actions": ["forget"]}}
			]}`,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewTerraformApprovalPlan([]byte(tc.planJSON)).IsNoop()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
