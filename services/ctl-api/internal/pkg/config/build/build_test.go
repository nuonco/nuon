package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func ptr[T any](v T) *T { return &v }

func role(name string) *config.AppAWSIAMRole {
	return &config.AppAWSIAMRole{
		Name:        name,
		DisplayName: name,
		Description: name,
		Policies: []config.AppAWSIAMPolicy{
			{Name: name + "-policy", GCPPermissions: []string{"compute.instances.get"}},
		},
	}
}

func rolesByType(roles []app.AppAWSIAMRoleConfig, typ app.AWSIAMRoleType) []string {
	var names []string
	for _, r := range roles {
		if r.Type == typ {
			names = append(names, r.Name)
		}
	}
	return names
}

func TestRunnerConfigKeepsPhoneHomeScriptURL(t *testing.T) {
	const scriptURL = "https://example.com/phonehome.py"

	obj := RunnerConfig(RunnerInputFromConfig(config.AppRunnerConfig{
		RunnerType:         string(app.AppRunnerTypeAWS),
		PhoneHomeScriptURL: scriptURL,
	}, "app1", "cfg1"))

	assert.Equal(t, scriptURL, obj.PhoneHomeScriptURL)
}

func TestPermissionsConfigKeepsCustomAndBreakGlassRoles(t *testing.T) {
	obj, err := PermissionsConfig(PermissionsInput{
		AppID:       "app1",
		AppConfigID: "cfg1",
		Permissions: &config.PermissionsConfig{
			ProvisionRole:   role("provision"),
			MaintenanceRole: role("maintenance"),
			DeprovisionRole: role("deprovision"),
			CustomRoles:     []*config.AppAWSIAMRole{role("custom-a"), role("custom-b")},
		},
		BreakGlassRoles: []*config.AppAWSIAMRole{role("break-glass")},
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"custom-a", "custom-b"}, rolesByType(obj.Roles, app.AWSIAMRoleTypeCustom))
	assert.Equal(t, []string{"break-glass"}, rolesByType(obj.Roles, app.AWSIAMRoleTypeBreakGlass))
	assert.Equal(t, []string{"provision"}, rolesByType(obj.Roles, app.AWSIAMRoleTypeRunnerProvision))

	for _, r := range obj.Roles {
		assert.Equal(t, "cfg1", r.AppConfigID)
		require.Len(t, r.Policies, 1)
		assert.Equal(t, []string{"compute.instances.get"}, r.Policies[0].GCPPermissions)
	}
}

func TestPermissionsConfigKeepsEnabledInStack(t *testing.T) {
	provision := role("provision")
	provision.EnabledInStack = ptr(false)

	obj, err := PermissionsConfig(PermissionsInput{
		AppConfigID: "cfg1",
		Permissions: &config.PermissionsConfig{
			ProvisionRole:   provision,
			MaintenanceRole: role("maintenance"),
			DeprovisionRole: role("deprovision"),
		},
	})
	require.NoError(t, err)

	found := false
	for _, r := range obj.Roles {
		if r.Type != app.AWSIAMRoleTypeRunnerProvision {
			continue
		}
		found = true
		assert.True(t, r.EnabledInStack.Valid)
		assert.False(t, r.EnabledInStack.Bool)
	}
	assert.True(t, found)
}

func TestPermissionsConfigRejectsMutuallyExclusivePolicy(t *testing.T) {
	custom := role("custom")
	custom.Policies[0].GCPPredefinedRole = "roles/viewer"

	_, err := PermissionsConfig(PermissionsInput{
		AppConfigID: "cfg1",
		Permissions: &config.PermissionsConfig{
			ProvisionRole:   role("provision"),
			MaintenanceRole: role("maintenance"),
			DeprovisionRole: role("deprovision"),
			CustomRoles:     []*config.AppAWSIAMRole{custom},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestBreakGlassConfigBuildsRoles(t *testing.T) {
	obj, err := BreakGlassConfig("app1", "cfg1", []*config.AppAWSIAMRole{role("bg")})
	require.NoError(t, err)

	require.Len(t, obj.Roles, 1)
	assert.Equal(t, app.AWSIAMRoleTypeBreakGlass, obj.Roles[0].Type)
	assert.Equal(t, "app1", obj.AppID)
	assert.Equal(t, "cfg1", obj.AppConfigID)
}

func TestSandboxConfigKeepsPulumiAndOperationRoles(t *testing.T) {
	obj, err := SandboxConfig(SandboxInputFromConfig(&config.AppSandboxConfig{
		Type:          config.AppSandboxTypePulumi,
		Runtime:       "go",
		PulumiVersion: "3.100.0",
		PulumiConfig:  map[string]string{"foo": "bar"},
		OperationRoles: []config.EntityOperationRole{
			{Operation: config.OperationType(app.OperationProvision), RoleName: "custom-a"},
		},
		SkipNoops:                    ptr(true),
		AutoApproveOnPoliciesPassing: ptr(true),
	}, "app1", "cfg1"))
	require.NoError(t, err)

	assert.Equal(t, config.AppSandboxTypePulumi, obj.Type)
	assert.Equal(t, "go", obj.Runtime)
	assert.Equal(t, "3.100.0", obj.PulumiVersion)
	require.Contains(t, obj.PulumiConfig, "foo")
	assert.Equal(t, "bar", *obj.PulumiConfig["foo"])
	require.Contains(t, obj.OperationRoles, string(app.OperationProvision))
	assert.Equal(t, "custom-a", *obj.OperationRoles[string(app.OperationProvision)])
	assert.True(t, *obj.SkipNoops)
	assert.True(t, *obj.AutoApproveOnPoliciesPassing)
}

func TestSandboxConfigDefaultsToTerraform(t *testing.T) {
	obj, err := SandboxConfig(SandboxInputFromConfig(&config.AppSandboxConfig{
		TerraformVersion: "1.9.0",
	}, "app1", "cfg1"))
	require.NoError(t, err)
	assert.Equal(t, config.AppSandboxTypeTerraform, obj.Type)
}

func TestSandboxConfigRejectsBadInput(t *testing.T) {
	for name, sandbox := range map[string]*config.AppSandboxConfig{
		"pulumi without runtime": {Type: config.AppSandboxTypePulumi},
		"pulumi bad runtime":     {Type: config.AppSandboxTypePulumi, Runtime: "cobol"},
		"terraform no version":   {Type: config.AppSandboxTypeTerraform},
		"unknown type":           {Type: "bicep"},
		"unknown operation": {
			TerraformVersion: "1.9.0",
			OperationRoles:   []config.EntityOperationRole{{Operation: "explode", RoleName: "r"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := SandboxConfig(SandboxInputFromConfig(sandbox, "app1", "cfg1"))
			require.Error(t, err)
		})
	}
}

func TestStackConfigKeepsCustomNestedStacksPending(t *testing.T) {
	obj, err := StackConfig(&config.StackConfig{
		Type:        string(app.StackTypeAWS),
		Name:        "stack",
		Description: "stack",
		CustomNestedStacks: []config.CustomNestedStack{
			{Name: "ns", TemplateURL: "https://example.com/t.yaml", Contents: "Resources: {}"},
		},
	}, "app1", "cfg1")
	require.NoError(t, err)

	require.Len(t, obj.CustomNestedStacks, 1)
	assert.Equal(t, config.CustomNestedStackStatusPending, obj.CustomNestedStacks[0].Status)
	assert.Equal(t, "ns", obj.CustomNestedStacks[0].Name)
}

func TestStackConfigDoesNotMutateInputStatus(t *testing.T) {
	in := &config.StackConfig{
		Type:        string(app.StackTypeAWS),
		Name:        "stack",
		Description: "stack",
		CustomNestedStacks: []config.CustomNestedStack{
			{Name: "ns", TemplateURL: "https://example.com/t.yaml", Contents: "Resources: {}"},
		},
	}
	_, err := StackConfig(in, "app1", "cfg1")
	require.NoError(t, err)
	assert.Empty(t, in.CustomNestedStacks[0].Status)
}

func TestStackConfigRejectsNestedStackWithoutContents(t *testing.T) {
	_, err := StackConfig(&config.StackConfig{
		Type:               string(app.StackTypeAWS),
		Name:               "stack",
		Description:        "stack",
		CustomNestedStacks: []config.CustomNestedStack{{Name: "ns", TemplateURL: "https://example.com/t.yaml"}},
	}, "app1", "cfg1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contents is required")
}

func TestPoliciesConfigKeepsName(t *testing.T) {
	obj, err := PoliciesConfig(PolicyInputsFromConfig(&config.PoliciesConfig{
		Policies: []config.AppPolicy{
			{Type: config.AppPolicyTypeSandbox, Name: "no-public-buckets", Contents: "package x"},
		},
	}), "app1", "cfg1")
	require.NoError(t, err)

	require.Len(t, obj.Policies, 1)
	assert.Equal(t, "no-public-buckets", obj.Policies[0].Name)
}

func TestPoliciesConfigRejectsUnknownType(t *testing.T) {
	_, err := PoliciesConfig([]PolicyInput{{Type: "nope", Contents: "package x"}}, "app1", "cfg1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid policy type")
}

func TestComponentConnectionKeepsToggles(t *testing.T) {
	comp := &config.Component{
		Name:           "api",
		Toggleable:     ptr(true),
		DefaultEnabled: ptr(false),
		HelmChart: &config.HelmChartComponentConfig{
			ChartName: "apichart",
			SkipNoops: ptr(true),
			Health: &config.ComponentHealthConfig{
				Enabled:        ptr(true),
				RequiredChecks: []string{"smoke"},
			},
		},
	}

	in, err := ComponentConnectionInputFromConfig(comp, "cmp1", "cfg1", nil)
	require.NoError(t, err)

	ccc, err := ComponentConnection(in)
	require.NoError(t, err)

	assert.True(t, *ccc.Toggleable)
	assert.False(t, *ccc.DefaultEnabled)
	assert.True(t, *ccc.SkipNoops)
	assert.True(t, *ccc.HealthEnabled)
	assert.Equal(t, app.ComponentHealthRequiredChecks{"smoke"}, ccc.HealthRequiredChecks)
}

func TestComponentConnectionKeepsKubernetesManifestDriftSchedule(t *testing.T) {
	comp := &config.Component{
		Name: "manifests",
		KubernetesManifest: &config.KubernetesManifestComponentConfig{
			Namespace:     "default",
			Manifest:      "kind: Namespace",
			DriftSchedule: ptr("0 * * * *"),
		},
	}

	in, err := ComponentConnectionInputFromConfig(comp, "cmp1", "cfg1", nil)
	require.NoError(t, err)

	ccc, err := ComponentConnection(in)
	require.NoError(t, err)
	assert.Equal(t, "0 * * * *", ccc.DriftSchedule)
}

func TestComponentConnectionRejectsUntypedComponent(t *testing.T) {
	_, err := ComponentConnectionInputFromConfig(&config.Component{Name: "orphan"}, "cmp1", "cfg1", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no type configuration")
}

func TestKubernetesManifestKeepsInlineManifest(t *testing.T) {
	cfg, err := KubernetesManifestComponentConfig(&config.KubernetesManifestComponentConfig{
		Namespace: "default",
		Manifest:  "kind: Namespace",
	}, VCS{})
	require.NoError(t, err)
	assert.Equal(t, "kind: Namespace", cfg.Manifest)
	assert.Nil(t, cfg.Kustomize)
}

func TestKubernetesManifestRejectsManifestAndKustomize(t *testing.T) {
	_, err := KubernetesManifestComponentConfig(&config.KubernetesManifestComponentConfig{
		Namespace: "default",
		Manifest:  "kind: Namespace",
		Kustomize: &config.KustomizeConfig{Path: "./overlays"},
	}, VCS{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one of")
}

func TestJobComponentConfigBuilds(t *testing.T) {
	cfg, err := JobComponentConfig(&config.JobComponentConfig{
		ImageURL:  "ghcr.io/acme/migrate",
		Tag:       "v1",
		Cmd:       []string{"/bin/migrate"},
		Args:      []string{"--up"},
		EnvVarMap: map[string]string{"LOG": "debug"},
		EnvVars:   []config.EnvironmentVariable{{Name: "REGION", Value: "us-west-2"}},
	})
	require.NoError(t, err)

	assert.Equal(t, "ghcr.io/acme/migrate", cfg.ImageURL)
	assert.Equal(t, "v1", cfg.Tag)
	assert.Equal(t, []string{"/bin/migrate"}, []string(cfg.Cmd))
	assert.Equal(t, []string{"--up"}, []string(cfg.Args))
	require.Contains(t, cfg.EnvVars, "LOG")
	require.Contains(t, cfg.EnvVars, "REGION")
}

func TestAttachTypeConfigSupportsJob(t *testing.T) {
	comp := &config.Component{
		Name: "migrate",
		Job:  &config.JobComponentConfig{ImageURL: "ghcr.io/acme/migrate", Tag: "v1"},
	}

	in, err := ComponentConnectionInputFromConfig(comp, "cmp1", "cfg1", nil)
	require.NoError(t, err)
	ccc, err := ComponentConnection(in)
	require.NoError(t, err)

	require.NoError(t, AttachTypeConfig(ccc, comp, VCS{}, ""))
	require.NotNil(t, ccc.JobComponentConfig)
	assert.Equal(t, "ghcr.io/acme/migrate", ccc.JobComponentConfig.ImageURL)
}

func TestExternalImageRejectsMultipleSources(t *testing.T) {
	_, err := ExternalImageComponentConfig(&config.ExternalImageComponentConfig{
		AWSECRImageConfig: &config.AWSECRConfig{ImageURL: "a", Tag: "1"},
		GCPGARImageConfig: &config.GCPGARConfig{ImageURL: "b", Tag: "1"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one")
}

func TestActionWorkflowConfigDefaultsEnableKubeConfig(t *testing.T) {
	awc := ActionWorkflowConfig(ActionWorkflowInput{
		ActionWorkflowID: "aw1",
		Role:             "custom-a",
	})
	assert.True(t, awc.EnableKubeConfig.Valid)
	assert.True(t, awc.EnableKubeConfig.Bool)
	assert.Equal(t, "custom-a", awc.Role)

	off := ActionWorkflowConfig(ActionWorkflowInput{EnableKubeConfig: ptr(false)})
	assert.True(t, off.EnableKubeConfig.Valid)
	assert.False(t, off.EnableKubeConfig.Bool)
}

func TestActionWorkflowConfigKeepsDependencyIDs(t *testing.T) {
	awc := ActionWorkflowConfig(ActionWorkflowInput{DependencyIDs: []string{"cmp1", "cmp2"}})
	assert.Equal(t, []string{"cmp1", "cmp2"}, []string(awc.ComponentDependencyIDs))
}

// Azure roles use an empty Statement list as a placeholder because the real
// grants come from RBAC; rejecting it broke real configs on the branch path.
func TestInlinePolicyAllowsEmptyStatementList(t *testing.T) {
	bg := role("break-glass")
	bg.Policies = []config.AppAWSIAMPolicy{
		{Name: "placeholder", Contents: `{"Version":"2012-10-17","Statement":[]}`},
	}

	_, err := BreakGlassConfig("app1", "cfg1", []*config.AppAWSIAMRole{bg})
	require.NoError(t, err)
}

func TestInlinePolicyStillRejectsMalformed(t *testing.T) {
	bad := role("bad")
	bad.Policies = []config.AppAWSIAMPolicy{{Name: "broken", Contents: `{"Statement": [`}}
	_, err := BreakGlassConfig("app1", "cfg1", []*config.AppAWSIAMRole{bad})
	require.Error(t, err)

	noEffect := role("no-effect")
	noEffect.Policies = []config.AppAWSIAMPolicy{
		{Name: "no-effect", Contents: `{"Statement":[{"Action":"s3:*"}]}`},
	}
	_, err = BreakGlassConfig("app1", "cfg1", []*config.AppAWSIAMRole{noEffect})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Effect")
}

// The checksum drives the sync's "did this component change" decision, so it
// must move for a change anywhere in the resolved component and stay put when
// nothing moved. Hashing only the component's own file would miss vars folded in
// from elsewhere.
func TestComponentChecksumTracksResolvedComponent(t *testing.T) {
	comp := func() *config.Component {
		return &config.Component{
			Name: "api",
			TerraformModule: &config.TerraformModuleComponentConfig{
				TerraformVersion: "1.9.0",
				EnvVarMap:        map[string]string{"REGION": "us-west-2"},
			},
		}
	}

	base, err := ComponentChecksum(comp())
	require.NoError(t, err)
	assert.NotEmpty(t, base)

	same, err := ComponentChecksum(comp())
	require.NoError(t, err)
	assert.Equal(t, base, same, "identical components must checksum the same")

	changed := comp()
	changed.TerraformModule.EnvVarMap["REGION"] = "us-east-1"
	changedSum, err := ComponentChecksum(changed)
	require.NoError(t, err)
	assert.NotEqual(t, base, changedSum, "a resolved value change must move the checksum")

	// The per-file Checksum field is excluded: it is an input to hashing, not
	// part of the component's meaning.
	withFileChecksum := comp()
	withFileChecksum.Checksum = "some-file-hash"
	fileSum, err := ComponentChecksum(withFileChecksum)
	require.NoError(t, err)
	assert.Equal(t, base, fileSum, "the per-file checksum must not affect the hash")
}
