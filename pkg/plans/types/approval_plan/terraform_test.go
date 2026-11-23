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
		"drift on attribute nothing references": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["update"],
					"before": {"endpoint": "db.example.com", "latest_restorable_time": "2026-11-23T17:17:39Z"},
					"after": {"endpoint": "db.example.com", "latest_restorable_time": "2026-11-23T17:57:40Z"}}}],
				"relevant_attributes": [{"resource": "aws_db_instance.this", "attribute": ["endpoint"]}]
			}`,
			want: true,
		},
		"drift without relevant attributes for the resource": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["update"],
					"before": {"latest_restorable_time": "2026-11-23T17:17:39Z"},
					"after": {"latest_restorable_time": "2026-11-23T17:57:40Z"}}}]
			}`,
			want: true,
		},
		"drift on referenced attribute": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["update"],
					"before": {"port": 5432, "latest_restorable_time": "2026-11-23T17:17:39Z"},
					"after": {"port": 5433, "latest_restorable_time": "2026-11-23T17:57:40Z"}}}],
				"relevant_attributes": [{"resource": "aws_db_instance.this", "attribute": ["port"]}]
			}`,
			want: false,
		},
		"drift on nested referenced attribute": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["update"],
					"before": {"tags": {"env": "prod"}},
					"after": {"tags": {"env": "dev"}}}}],
				"relevant_attributes": [{"resource": "aws_db_instance.this", "attribute": ["tags", "env"]}]
			}`,
			want: false,
		},
		"drift on sibling of nested referenced attribute": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["update"],
					"before": {"tags": {"env": "prod", "last_backup": "2026-11-23T17:17:39Z"}},
					"after": {"tags": {"env": "prod", "last_backup": "2026-11-23T17:57:40Z"}}}}],
				"relevant_attributes": [{"resource": "aws_db_instance.this", "attribute": ["tags", "env"]}]
			}`,
			want: true,
		},
		"drift on referenced list element": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["update"],
					"before": {"endpoint": [{"port": 5432, "status": "a"}, {"port": 5432}]},
					"after": {"endpoint": [{"port": 5432, "status": "b"}, {"port": 5433}]}}}],
				"relevant_attributes": [{"resource": "aws_db_instance.this", "attribute": ["endpoint", 1, "port"]}]
			}`,
			want: false,
		},
		"drift on sibling of referenced list element": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["update"],
					"before": {"endpoint": [{"port": 5432, "status": "a"}]},
					"after": {"endpoint": [{"port": 5432, "status": "b"}]}}}],
				"relevant_attributes": [{"resource": "aws_db_instance.this", "attribute": ["endpoint", 0, "port"]}]
			}`,
			want: true,
		},
		"resource replaced outside terraform": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["delete", "create"],
					"before": {"port": 5432}, "after": {"port": 5432}}}]
			}`,
			want: false,
		},
		"drift on resource referenced as a whole": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_cloudwatch_log_group.this", "change": {"actions": ["update"],
					"before": {"retention_in_days": 7},
					"after": {"retention_in_days": 30}}}],
				"relevant_attributes": [{"resource": "aws_cloudwatch_log_group.this", "attribute": []}]
			}`,
			want: false,
		},
		"resource deleted outside terraform": {
			planJSON: `{"format_version": "1.2",
				"resource_changes": [{"address": "aws_db_instance.this", "change": {"actions": ["no-op"]}}],
				"resource_drift": [{"address": "aws_db_instance.this", "change": {"actions": ["delete"],
					"before": {"port": 5432}, "after": null}}]
			}`,
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
