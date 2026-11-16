package cloudformation

import (
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/eks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func eksAccessTestInput(clusterName string, st *state.State) *stacks.TemplateInput {
	inp := rolesTestInput()
	inp.InstallState = st
	inp.AppCfg.BreakGlassConfig.Roles[0].EKSAccess = &app.AppAWSIAMRoleEKSAccess{
		ClusterName:      clusterName,
		KubernetesGroups: []string{"maintenance"},
		Policies: []app.AppAWSIAMRoleEKSAccessPolicy{{
			PolicyARN:   "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
			AccessScope: "namespace",
			Namespaces:  []string{"default"},
		}},
	}
	inp.AppCfg.PermissionsConfig.Roles = append(inp.AppCfg.PermissionsConfig.Roles, inp.AppCfg.BreakGlassConfig.Roles[0])
	return inp
}

func TestGetEKSAccessResources(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := eksAccessTestInput("acme-cluster", state.New())

	rsrcs := tpl.getEKSAccessResources(inp, tagBuilder{installID: "inl123"})
	require.Len(t, rsrcs, 1, "a role listed under both permissions and break glass gets one entry")

	entry, ok := rsrcs["BreakGlassRoleEKSAccessEntry"].(*eks.AccessEntry)
	require.True(t, ok)
	assert.Equal(t, "BreakGlassRoleEKSAccess", entry.AWSCloudFormationCondition)
	assert.Equal(t, cloudformation.Ref("BreakGlassRoleEKSClusterName"), entry.ClusterName)
	assert.Equal(t, cloudformation.GetAtt("BreakGlassRole", "Arn"), entry.PrincipalArn)
	assert.Equal(t, []string{"maintenance"}, entry.KubernetesGroups)
	require.Len(t, entry.AccessPolicies, 1)
	assert.Equal(t, "namespace", entry.AccessPolicies[0].AccessScope.Type)
	assert.Equal(t, []string{"default"}, entry.AccessPolicies[0].AccessScope.Namespaces)

	conditions := tpl.getEKSAccessConditions(inp)
	assert.Equal(t, cloudformation.And([]string{
		cloudformation.Equals(cloudformation.Ref("BreakGlassRoleEnabled"), "true"),
		cloudformation.Not([]string{cloudformation.Equals(cloudformation.Ref("BreakGlassRoleEKSClusterName"), "")}),
	}), conditions["BreakGlassRoleEKSAccess"])
}

func TestGetEKSAccessParametersClusterNameDefault(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	withSandbox := &state.State{Sandbox: &state.SandboxState{Outputs: map[string]any{
		"cluster": map[string]any{"name": "acme-eks"},
	}}}
	withInputs := &state.State{Inputs: &state.InputsState{Inputs: map[string]string{"cluster_name": "acme-input"}}}

	tests := []struct {
		name        string
		clusterName string
		state       *state.State
		want        string
	}{
		{"literal", "acme-cluster", state.New(), "acme-cluster"},
		{"sandbox output", "{{.nuon.sandbox.outputs.cluster.name}}", withSandbox, "acme-eks"},
		{"sandbox output before the sandbox exists", "{{.nuon.sandbox.outputs.cluster.name}}", state.New(), ""},
		{"install input", "{{.nuon.inputs.inputs.cluster_name}}", withInputs, "acme-input"},
		{"no install state", "{{.nuon.sandbox.outputs.cluster.name}}", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tpl.getEKSAccessParameters(eksAccessTestInput(tt.clusterName, tt.state))
			param, ok := params["BreakGlassRoleEKSClusterName"]
			require.True(t, ok)
			assert.Equal(t, tt.want, param.Default)
		})
	}
}
