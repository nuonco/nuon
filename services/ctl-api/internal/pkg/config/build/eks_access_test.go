package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func eksAccessRole(name string, access *config.AppAWSIAMRoleEKSAccess) *config.AppAWSIAMRole {
	r := role(name)
	r.CloudPlatform = "aws"
	r.Policies = []config.AppAWSIAMPolicy{{
		Name:     name + "-policy",
		Contents: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"eks:DescribeCluster","Resource":"*"}]}`,
	}}
	r.EKSAccess = access
	return r
}

func validEKSAccess() *config.AppAWSIAMRoleEKSAccess {
	return &config.AppAWSIAMRoleEKSAccess{
		ClusterName:      "{{.nuon.inputs.inputs.cluster_name}}",
		KubernetesGroups: []string{"maintenance"},
		Policies: []config.AppAWSIAMRoleEKSAccessPolicy{{
			PolicyARN: "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
		}},
	}
}

func TestBreakGlassConfigEKSAccess(t *testing.T) {
	tests := []struct {
		name      string
		stackType string
		mutate    func(*config.AppAWSIAMRoleEKSAccess)
		wantErr   string
	}{
		{name: "valid", stackType: "aws-cloudformation"},
		{name: "non cloudformation stack", stackType: "aws-terraform", wantErr: `eks_access requires stack type "aws-cloudformation"`},
		{name: "no stack", stackType: "", wantErr: `eks_access requires stack type "aws-cloudformation"`},
		{
			name:      "system group",
			stackType: "aws-cloudformation",
			mutate:    func(a *config.AppAWSIAMRoleEKSAccess) { a.KubernetesGroups = []string{"system:masters"} },
			wantErr:   `kubernetes group "system:masters" is reserved`,
		},
		{
			name:      "empty group",
			stackType: "aws-cloudformation",
			mutate:    func(a *config.AppAWSIAMRoleEKSAccess) { a.KubernetesGroups = []string{" "} },
			wantErr:   "kubernetes_groups has an empty group",
		},
		{
			name:      "missing cluster name",
			stackType: "aws-cloudformation",
			mutate:    func(a *config.AppAWSIAMRoleEKSAccess) { a.ClusterName = "" },
			wantErr:   "eks_access.cluster_name is required",
		},
		{
			name:      "bad policy arn",
			stackType: "aws-cloudformation",
			mutate: func(a *config.AppAWSIAMRoleEKSAccess) {
				a.Policies[0].PolicyARN = "arn:aws:iam::aws:policy/AdministratorAccess"
			},
			wantErr: "is not an EKS cluster access policy ARN",
		},
		{
			name:      "namespace scope without namespaces",
			stackType: "aws-cloudformation",
			mutate:    func(a *config.AppAWSIAMRoleEKSAccess) { a.Policies[0].AccessScope = config.EKSAccessScopeNamespace },
			wantErr:   "has access_scope namespace but no namespaces",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access := validEKSAccess()
			if tt.mutate != nil {
				tt.mutate(access)
			}

			obj, err := BreakGlassConfig("app1", "cfg1", tt.stackType, []*config.AppAWSIAMRole{eksAccessRole("bg", access)})
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, obj.Roles, 1)
			require.NotNil(t, obj.Roles[0].EKSAccess)
			assert.Equal(t, config.EKSAccessScopeCluster, obj.Roles[0].EKSAccess.Policies[0].AccessScope)
		})
	}
}

func TestBreakGlassConfigWithoutEKSAccessIgnoresStackType(t *testing.T) {
	_, err := BreakGlassConfig("app1", "cfg1", "aws-terraform", []*config.AppAWSIAMRole{role("bg")})
	require.NoError(t, err)
}

func TestPermissionsConfigRejectsEKSAccessOnRunnerRoles(t *testing.T) {
	_, err := PermissionsConfig(PermissionsInput{
		AppConfigID: "cfg1",
		StackType:   "aws-cloudformation",
		Permissions: &config.PermissionsConfig{
			ProvisionRole:   eksAccessRole("provision", validEKSAccess()),
			MaintenanceRole: role("maintenance"),
			DeprovisionRole: role("deprovision"),
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only supported on break glass and custom roles")
}
