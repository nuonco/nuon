package componentteardownsyncandplan

import (
	"testing"

	"github.com/stretchr/testify/require"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

func TestCheckTeardownPlan(t *testing.T) {
	retainPlan := &plantypes.DeployPlan{Teardown: &plantypes.TeardownPlan{Retain: []string{"aws_rds_cluster.main"}}}

	tests := map[string]struct {
		deployPlan      *plantypes.DeployPlan
		compType        app.ComponentType
		sandboxMode     bool
		contentsDisplay string
		wantErr         string
	}{
		"plan retains resources": {
			deployPlan:      retainPlan,
			compType:        app.ComponentTypeTerraformModule,
			contentsDisplay: planForgettingRetained,
		},
		"plan deletes a retained resource": {
			deployPlan:      retainPlan,
			compType:        app.ComponentTypeTerraformModule,
			contentsDisplay: planDeletingRetained,
			wantErr:         "update it and retry",
		},
		"no teardown settings": {
			deployPlan:      &plantypes.DeployPlan{},
			compType:        app.ComponentTypeTerraformModule,
			contentsDisplay: planDeletingRetained,
		},
		"empty retain list": {
			deployPlan:      &plantypes.DeployPlan{Teardown: &plantypes.TeardownPlan{}},
			compType:        app.ComponentTypeTerraformModule,
			contentsDisplay: planDeletingRetained,
		},
		"non-terraform component": {
			deployPlan:      retainPlan,
			compType:        app.ComponentTypeHelmChart,
			contentsDisplay: `{"op": "helm"}`,
		},
		"sandbox mode plans are fake": {
			deployPlan:      retainPlan,
			compType:        app.ComponentTypeTerraformModule,
			sandboxMode:     true,
			contentsDisplay: planDeletingRetained,
		},
		"unparseable plan": {
			deployPlan:      retainPlan,
			compType:        app.ComponentTypeTerraformModule,
			contentsDisplay: "not json",
			wantErr:         "unable to parse teardown plan",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkTeardownPlan(tc.deployPlan, tc.compType, tc.sandboxMode, tc.contentsDisplay)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
