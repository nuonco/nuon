package approvalplan

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpentofuApprovalPlan_IsNoop(t *testing.T) {
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
		"create": {
			planJSON: `{"format_version": "1.2", "resource_changes": [
				{"address": "aws_rds_cluster.main", "mode": "managed", "change": {"actions": ["create"]}}
			]}`,
			want: false,
		},
		"replace is not noop": {
			planJSON: `{"format_version": "1.2", "resource_changes": [
				{"address": "aws_rds_cluster.main", "mode": "managed", "change": {"actions": ["delete", "create"]}}
			]}`,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewOpentofuApprovalPlan([]byte(tc.planJSON)).IsNoop()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
