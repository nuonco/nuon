package activities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	planDeletingRetained = `{
  "format_version": "1.2",
  "resource_changes": [
    {"address": "aws_db_subnet_group.main", "mode": "managed", "type": "aws_db_subnet_group", "name": "main", "change": {"actions": ["delete"]}},
    {"address": "aws_rds_cluster.main", "mode": "managed", "type": "aws_rds_cluster", "name": "main", "change": {"actions": ["delete"]}}
  ]
}`
	planForgettingRetained = `{
  "format_version": "1.2",
  "resource_changes": [
    {"address": "aws_db_subnet_group.main", "mode": "managed", "type": "aws_db_subnet_group", "name": "main", "change": {"actions": ["delete"]}},
    {"address": "aws_rds_cluster.main", "mode": "managed", "type": "aws_rds_cluster", "name": "main", "change": {"actions": ["forget"]}}
  ]
}`
)

func TestCheckTeardownPlanContents(t *testing.T) {
	retain := []string{"aws_rds_cluster.main"}

	tests := map[string]struct {
		contents string
		wantErr  string
	}{
		"plan retains resources": {
			contents: planForgettingRetained,
		},
		"plan deletes a retained resource": {
			contents: planDeletingRetained,
			wantErr:  "update it and retry",
		},
		"unparseable plan": {
			contents: "not json",
			wantErr:  "unable to parse teardown plan",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkTeardownPlanContents([]byte(tc.contents), retain)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
