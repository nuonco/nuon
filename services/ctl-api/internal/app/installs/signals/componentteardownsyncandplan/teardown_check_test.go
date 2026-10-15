package componentteardownsyncandplan

import (
	"testing"

	"github.com/stretchr/testify/require"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestTeardownPlanRetain(t *testing.T) {
	retainPlan := &plantypes.DeployPlan{Teardown: &plantypes.TeardownPlan{Retain: []string{"aws_rds_cluster.main"}}}

	tests := map[string]struct {
		deployPlan  *plantypes.DeployPlan
		compType    app.ComponentType
		sandboxMode bool
		want        []string
	}{
		"plan retains resources": {
			deployPlan: retainPlan,
			compType:   app.ComponentTypeTerraformModule,
			want:       []string{"aws_rds_cluster.main"},
		},
		"no teardown settings": {
			deployPlan: &plantypes.DeployPlan{},
			compType:   app.ComponentTypeTerraformModule,
		},
		"empty retain list": {
			deployPlan: &plantypes.DeployPlan{Teardown: &plantypes.TeardownPlan{}},
			compType:   app.ComponentTypeTerraformModule,
		},
		"non-terraform component": {
			deployPlan: retainPlan,
			compType:   app.ComponentTypeHelmChart,
		},
		"sandbox mode plans are fake": {
			deployPlan:  retainPlan,
			compType:    app.ComponentTypeTerraformModule,
			sandboxMode: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, teardownPlanRetain(tc.deployPlan, tc.compType, tc.sandboxMode))
		})
	}
}
