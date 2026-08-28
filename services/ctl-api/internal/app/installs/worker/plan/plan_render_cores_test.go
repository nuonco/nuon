package plan

import (
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func hstore(m map[string]string) pgtype.Hstore {
	out := make(pgtype.Hstore, len(m))
	for k, v := range m {
		v := v
		out[k] = &v
	}
	return out
}

func awsStack(region string) *app.InstallStack {
	return &app.InstallStack{
		InstallStackOutputs: app.InstallStackOutputs{
			AWSStackOutputs: &app.AWSStackOutputs{Region: region},
		},
	}
}

func TestRenderTerraformDeployPlan(t *testing.T) {
	l := zap.NewNop()
	p := &Planner{}

	st := &state.State{
		Inputs: &state.InputsState{
			Inputs: map[string]string{
				config.TFVarsOverrideInputName("mycomp"): `key = "{{.nuon.install.id}}"`,
			},
		},
	}
	stateData := map[string]any{
		"install": map[string]any{"id": "inst123"},
	}

	clusterInfo := &kube.ClusterInfo{ID: "cluster-1"}
	var gotAuth *CloudAuth

	plan, err := p.RenderTerraformDeployPlan(l, &RenderTerraformDeployPlanInput{
		Stack:     awsStack("us-west-2"),
		State:     st,
		StateData: stateData,
		InstallDeploy: &app.InstallDeploy{
			ID:            "dep123",
			ComponentName: "mycomp",
		},
		CompBuild: &app.ComponentBuild{
			ComponentConfigConnection: app.ComponentConfigConnection{
				TerraformModuleComponentConfig: &app.TerraformModuleComponentConfig{
					Variables:      hstore(map[string]string{"install_id": "{{.nuon.install.id}}"}),
					EnvVars:        hstore(map[string]string{"FOO": "bar"}),
					VariablesFiles: []string{"base.tfvars"},
				},
			},
		},
		WorkspaceID:   "ws123",
		RoleSelection: &operationroles.RoleSelection{RoleARN: "arn:aws:iam::123:role/deploy"},
		ResolveClusterInfo: func(auth *CloudAuth) (*kube.ClusterInfo, error) {
			gotAuth = auth
			return clusterInfo, nil
		},
	})
	require.NoError(t, err)

	assert.Equal(t, map[string]any{"install_id": "inst123"}, plan.Vars)
	assert.Equal(t, map[string]string{"FOO": "bar"}, plan.EnvVars)
	assert.Equal(t, []string{"base.tfvars", `key = "inst123"`}, plan.VarsFiles)
	assert.Same(t, st, plan.State)
	assert.Equal(t, "ws123", plan.TerraformBackend.WorkspaceID)
	assert.Same(t, clusterInfo, plan.ClusterInfo)
	assert.False(t, plan.Hooks.Enabled)

	require.NotNil(t, plan.AWSAuth)
	assert.Equal(t, "us-west-2", plan.AWSAuth.Region)
	assert.Equal(t, "arn:aws:iam::123:role/deploy", plan.AWSAuth.AssumeRole.RoleARN)
	assert.Equal(t, "component-deploy-dep123", plan.AWSAuth.AssumeRole.SessionName)
	require.NotNil(t, gotAuth)
	assert.Same(t, plan.AWSAuth, gotAuth.AWS)
	assert.Nil(t, plan.AzureAuth)
	assert.Nil(t, plan.GCPAuth)
}

func TestRenderTerraformDeployPlan_MissingRole(t *testing.T) {
	p := &Planner{}
	_, err := p.RenderTerraformDeployPlan(zap.NewNop(), &RenderTerraformDeployPlanInput{
		Stack:         awsStack("us-west-2"),
		State:         &state.State{},
		StateData:     map[string]any{},
		InstallDeploy: &app.InstallDeploy{ID: "dep123", ComponentName: "mycomp"},
		CompBuild: &app.ComponentBuild{
			ComponentConfigConnection: app.ComponentConfigConnection{
				TerraformModuleComponentConfig: &app.TerraformModuleComponentConfig{},
			},
		},
		RoleSelection: &operationroles.RoleSelection{},
	})
	require.ErrorContains(t, err, "unable to get auth for deploy")
}

func TestDeploySrcRef(t *testing.T) {
	deploy := &app.InstallDeploy{ID: "dep123", ComponentBuildID: "bld123"}

	srcTag, srcDigest := DeploySrcRef(deploy, &app.ComponentBuild{})
	assert.Equal(t, "bld123", srcTag)
	assert.Empty(t, srcDigest)

	srcTag, srcDigest = DeploySrcRef(deploy, &app.ComponentBuild{SourceDigest: "sha256:abc"})
	assert.Equal(t, "sha256:abc", srcTag)
	assert.Equal(t, "sha256:abc", srcDigest)
}

func TestRenderSyncOCIPlan(t *testing.T) {
	l := zap.NewNop()
	p := &Planner{}

	srcCfg := &configs.OCIRegistryRepository{}

	tests := []struct {
		name        string
		compBuild   *app.ComponentBuild
		dstCfg      *configs.OCIRegistryRepository
		sandboxMode bool
		wantSrcTag  string
		wantDstTag  string
	}{
		{
			name:       "no_digest_uses_ids",
			compBuild:  &app.ComponentBuild{ID: "bld123"},
			dstCfg:     &configs.OCIRegistryRepository{},
			wantSrcTag: "bld-from-deploy",
			wantDstTag: "bld-from-deploy",
		},
		{
			name: "digest_without_resolved_tag_uses_build_id",
			compBuild: &app.ComponentBuild{
				ID:           "bld123",
				SourceDigest: "sha256:abc",
			},
			dstCfg:     &configs.OCIRegistryRepository{},
			wantSrcTag: "sha256:abc",
			wantDstTag: "bld123",
		},
		{
			name: "digest_with_resolved_tag",
			compBuild: &app.ComponentBuild{
				ID:           "bld123",
				SourceDigest: "sha256:abc",
				ResolvedTag:  "1.25.5",
			},
			dstCfg:     &configs.OCIRegistryRepository{},
			wantSrcTag: "sha256:abc",
			wantDstTag: "1.25.5",
		},
		{
			name: "ecr_prefixes_component_name",
			compBuild: &app.ComponentBuild{
				ID:           "bld123",
				SourceDigest: "sha256:abc",
				ResolvedTag:  "1.25.5",
			},
			dstCfg:     &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypeECR},
			wantSrcTag: "sha256:abc",
			wantDstTag: "my-comp-1.25.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pln, err := p.RenderSyncOCIPlan(l, &RenderSyncOCIPlanInput{
				Deploy: &app.InstallDeploy{
					ID:               "dep123",
					ComponentBuildID: "bld-from-deploy",
					ComponentName:    "My Comp",
				},
				CompBuild: tt.compBuild,
				Install: &app.Install{
					SandboxMode: sql.NullBool{Bool: tt.sandboxMode, Valid: true},
				},
				SrcCfg: srcCfg,
				DstCfg: tt.dstCfg,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantSrcTag, pln.SrcTag)
			assert.Equal(t, tt.wantDstTag, pln.DstTag)
			assert.Same(t, srcCfg, pln.Src)
			assert.Same(t, tt.dstCfg, pln.Dst)
			if tt.sandboxMode {
				require.NotNil(t, pln.SandboxMode)
				assert.True(t, pln.SandboxMode.Enabled)
			} else {
				assert.Nil(t, pln.SandboxMode)
			}
		})
	}
}

func TestResolveKubernetesContextFromData_SandboxDefault(t *testing.T) {
	p := &Planner{}
	cloudAuth := &CloudAuth{AWS: &awscredentials.Config{Region: "us-west-2"}}

	stateData := map[string]any{
		"sandbox": map[string]any{
			"outputs": map[string]any{
				"cluster": map[string]any{
					"name":                       "my-cluster",
					"endpoint":                   "https://eks.example.com",
					"certificate_authority_data": "cadata",
				},
			},
		},
	}

	info, err := p.ResolveKubernetesContextFromData(zap.NewNop(), "", &app.AppConfig{}, awsStack("us-west-2"), stateData, cloudAuth)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "my-cluster", info.ID)
	assert.Equal(t, "https://eks.example.com", info.Endpoint)
	assert.Equal(t, "cadata", info.CAData)
	assert.Same(t, cloudAuth.AWS, info.AWSAuth)
}

func TestResolveKubernetesContextFromData_NoClusterOutputs(t *testing.T) {
	p := &Planner{}

	info, err := p.ResolveKubernetesContextFromData(
		zap.NewNop(), "", &app.AppConfig{}, awsStack("us-west-2"),
		map[string]any{"sandbox": map[string]any{"outputs": map[string]any{}}},
		&CloudAuth{AWS: &awscredentials.Config{}},
	)
	require.NoError(t, err)
	assert.Nil(t, info)
}
