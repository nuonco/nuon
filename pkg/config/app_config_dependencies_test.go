package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	configdiff "github.com/nuonco/nuon/pkg/config/diff"
)

func TestAppConfigDiffPropagatesRoleChanges(t *testing.T) {
	oldCfg := appConfigWithRoleDependency(`{"Statement":[{"Action":"s3:GetObject"}]}`)
	newCfg := appConfigWithRoleDependency(`{"Statement":[{"Action":"s3:*"}]}`)

	result := newCfg.Diff(oldCfg)

	role := findResourceDiff(result, RoleResourceID("maintenance"))
	require.NotNil(t, role)
	require.True(t, role.DirectSummary().HasChanged)

	stack := findResourceDiff(result, StackResourceID)
	require.NotNil(t, stack)
	require.False(t, stack.DirectSummary().HasChanged)
	require.True(t, stack.Impacted)
	require.Contains(t, stack.ImpactReasons, configdiff.ImpactReason{
		From: RoleResourceID("maintenance"),
		Edge: configdiff.EdgeReasonStackRender,
	})

	component := findResourceDiff(result, ComponentResourceID("api"))
	require.NotNil(t, component)
	require.False(t, component.DirectSummary().HasChanged)
	require.True(t, component.Impacted)
	require.Contains(t, component.ImpactReasons, configdiff.ImpactReason{
		From: RoleResourceID("maintenance"),
		Edge: configdiff.EdgeReasonOperationRole,
	})
}

func TestAppConfigDiffPropagatesTransitiveComponentChanges(t *testing.T) {
	oldCfg := &AppConfig{Components: ComponentList{
		{Name: "database", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.5.0"}},
		{Name: "api", Type: HelmChartComponentType, Dependencies: []string{"database"}, HelmChart: &HelmChartComponentConfig{}},
		{Name: "web", Type: HelmChartComponentType, Dependencies: []string{"api"}, HelmChart: &HelmChartComponentConfig{}},
	}}
	newCfg := &AppConfig{Components: ComponentList{
		{Name: "database", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}},
		{Name: "api", Type: HelmChartComponentType, Dependencies: []string{"database"}, HelmChart: &HelmChartComponentConfig{}},
		{Name: "web", Type: HelmChartComponentType, Dependencies: []string{"api"}, HelmChart: &HelmChartComponentConfig{}},
	}}

	result := newCfg.Diff(oldCfg)

	api := findResourceDiff(result, ComponentResourceID("api"))
	require.False(t, api.DirectSummary().HasChanged)
	require.True(t, api.Impacted)
	web := findResourceDiff(result, ComponentResourceID("web"))
	require.False(t, web.DirectSummary().HasChanged)
	require.True(t, web.Impacted)
	require.Contains(t, web.ImpactReasons, configdiff.ImpactReason{
		From: ComponentResourceID("api"),
		Edge: configdiff.EdgeReasonComponentDependency,
	})
}

func TestAppConfigDiffPropagatesNamedPolicyChanges(t *testing.T) {
	oldCfg := appConfigWithNamedPolicy(nil)
	newCfg := appConfigWithNamedPolicy([]NamedIAMPolicy{{
		Name:     "sample-logs",
		Contents: `{"Statement":[{"Action":"logs:GetLogEvents"}]}`,
	}})

	result := newCfg.Diff(oldCfg)

	policy := findResourceDiff(result, NamedPolicyResourceID("sample-logs"))
	require.NotNil(t, policy)
	require.True(t, policy.DirectSummary().HasChanged)

	stack := findResourceDiff(result, StackResourceID)
	require.NotNil(t, stack)
	require.False(t, stack.DirectSummary().HasChanged)
	require.True(t, stack.Impacted)
	require.Contains(t, stack.ImpactReasons, configdiff.ImpactReason{
		From: NamedPolicyResourceID("sample-logs"),
		Edge: configdiff.EdgeReasonStackRender,
	})
}

func TestAppConfigDiff_DependencyOrderProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		dependencies := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 0, 10)
		makeConfig := func(dependencies []string) *AppConfig {
			return &AppConfig{Components: ComponentList{{Name: "consumer", Type: TerraformModuleComponentType, Dependencies: dependencies, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}}}}
		}
		old := makeConfig(drawPermutation(t, dependencies))
		new := makeConfig(append([]string{}, drawPermutation(t, dependencies)...))
		oldSnapshot, err := json.Marshal(old)
		require.NoError(t, err)
		newSnapshot, err := json.Marshal(new)
		require.NoError(t, err)
		require.False(t, new.Diff(old).Summary().HasChanged)
		require.Empty(t, new.Diff(old).FormatChanged(""))
		require.False(t, old.Diff(new).Summary().HasChanged)
		oldAfter, err := json.Marshal(old)
		require.NoError(t, err)
		newAfter, err := json.Marshal(new)
		require.NoError(t, err)
		require.Equal(t, oldSnapshot, oldAfter)
		require.Equal(t, newSnapshot, newAfter)

		added := makeConfig(append(append([]string{}, dependencies...), "extra_dependency"))
		require.True(t, added.Diff(old).Summary().HasChanged, "new dependencies must remain visible")
		require.True(t, old.Diff(added).Summary().HasChanged, "removed dependencies must remain visible")
		if len(dependencies) > 0 {
			replaced := makeConfig(append([]string{"replacement_dependency"}, dependencies[1:]...))
			require.True(t, replaced.Diff(old).Summary().HasChanged, "same-length membership changes must remain visible")
			require.True(t, old.Diff(replaced).Summary().HasChanged)
		}
	})
}

func TestAppConfigDiff_ImpactReachabilityProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		names := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 2, 8)
		old := &AppConfig{}
		reachable := make([][]bool, len(names))
		for i, name := range names {
			reachable[i] = make([]bool, len(names))
			old.Components = append(old.Components, &Component{Name: name, Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}})
			old.Actions = append(old.Actions, &ActionConfig{Name: "inspect_" + name, Dependencies: []string{name}, Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo done"}}})
		}
		for dependency := range names {
			for dependent := range names {
				if dependency != dependent && hegel.Draw(t, hegel.Booleans()) {
					old.Components[dependent].Dependencies = append(old.Components[dependent].Dependencies, names[dependency])
					reachable[dependency][dependent] = true
				}
			}
		}
		for k := range names {
			for i := range names {
				for j := range names {
					reachable[i][j] = reachable[i][j] || reachable[i][k] && reachable[k][j]
				}
			}
		}
		old.Components = append(old.Components, &Component{Name: "isolated_component", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}})
		before, err := json.Marshal(old)
		require.NoError(t, err)
		var new AppConfig
		require.NoError(t, json.Unmarshal(before, &new))
		root := hegel.Draw(t, hegel.Integers(0, len(names)-1))
		new.Components[root].TerraformModule.TerraformVersion = "1.7.0"
		result := new.Diff(old)
		for i, name := range names {
			component := result.FindResource(ComponentResourceID(name))
			require.NotNil(t, component)
			require.Equal(t, i == root, component.DirectSummary().HasChanged, "direct change for %s", name)
			require.Equal(t, i != root && reachable[root][i], component.Impacted, "impact for %s", name)
			action := result.FindResource(ActionResourceID("inspect_" + name))
			require.NotNil(t, action)
			require.False(t, action.DirectSummary().HasChanged)
			require.Equal(t, i == root || reachable[root][i], action.Impacted, "action impact for %s", name)
		}
		isolated := result.FindResource(ComponentResourceID("isolated_component"))
		require.False(t, isolated.Summary().HasChanged)
		require.False(t, old.Diff(old).Summary().HasChanged, "impact state must not leak into later comparisons")
		require.Equal(t, result, new.Diff(old))
		after, err := json.Marshal(old)
		require.NoError(t, err)
		require.Equal(t, before, after)
	})
}

func TestAppConfigDiff_ReferenceExtractionProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		name := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		for _, source := range []struct {
			expression string
			id         configdiff.NodeID
			reason     configdiff.EdgeReason
		}{
			{".nuon.inputs.inputs." + name, InputResourceID(name), configdiff.EdgeReasonInputReference},
			{".nuon.install.inputs." + name, InputResourceID(name), configdiff.EdgeReasonInputReference},
			{".nuon.secrets." + name, SecretResourceID(name), configdiff.EdgeReasonSecretReference},
			{".nuon.sandbox.outputs.region", SandboxResourceID, configdiff.EdgeReasonSandboxOutput},
			{".nuon.install.sandbox.outputs.region", SandboxResourceID, configdiff.EdgeReasonSandboxOutput},
			{".nuon.install_stack.outputs.region", StackResourceID, configdiff.EdgeReasonInstallStackOutput},
			{".nuon.components.source_" + name + ".outputs.region", ComponentResourceID("source_" + name), configdiff.EdgeReasonComponentReference},
			{".nuon.install.components.source_" + name + ".outputs.region", ComponentResourceID("source_" + name), configdiff.EdgeReasonComponentReference},
		} {
			cfg := baseConfig()
			cfg.Components = ComponentList{
				{Name: "source_" + name, Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}},
				{Name: "consumer", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{VarsMap: map[string]string{"region": "{{ " + source.expression + " }}"}}},
				{Name: "transitive", Type: TerraformModuleComponentType, Dependencies: []string{"consumer"}, TerraformModule: &TerraformModuleComponentConfig{}},
				{Name: "isolated", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{}},
			}
			cfg.Actions = []*ActionConfig{{Name: "consumer", Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo {{ " + source.expression + " }}"}}}}
			before, err := json.Marshal(cfg)
			require.NoError(t, err)
			result := cfg.Diff(cfg, source.id)
			require.False(t, result.DirectSummary().HasChanged)
			for _, id := range []configdiff.NodeID{ComponentResourceID("consumer"), ActionResourceID("consumer")} {
				consumer := result.FindResource(id)
				require.True(t, consumer.Impacted, "%s must consume %s", id, source.expression)
				require.Equal(t, []configdiff.ImpactReason{{From: source.id, Edge: source.reason}}, consumer.ImpactReasons)
			}
			require.True(t, result.FindResource(ComponentResourceID("transitive")).Impacted)
			require.False(t, result.FindResource(ComponentResourceID("isolated")).Summary().HasChanged)
			after, err := json.Marshal(cfg)
			require.NoError(t, err)
			require.Equal(t, before, after)
		}
	})
}

func TestAppConfigDiff_OldAndNewEdgeUnionProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		name := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		for _, kind := range []string{"dependency", "reference"} {
			for _, oldEdge := range []bool{false, true} {
				for _, newEdge := range []bool{false, true} {
					makeConfig := func(edge bool) *AppConfig {
						cfg := &AppConfig{Components: ComponentList{
							{Name: "source_" + name, Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}},
							{Name: "consumer", Type: TerraformModuleComponentType, TerraformModule: &TerraformModuleComponentConfig{}},
							{Name: "downstream", Type: JobComponentType, Dependencies: []string{"consumer"}, Job: &JobComponentConfig{}},
							{Name: "isolated", Type: JobComponentType, Job: &JobComponentConfig{}},
						}}
						if edge {
							if kind == "dependency" {
								cfg.Components[1].Dependencies = []string{"source_" + name}
							} else {
								cfg.Components[1].TerraformModule.VarsMap = map[string]string{"region": "{{ .nuon.components.source_" + name + ".outputs.region }}"}
							}
						}
						return cfg
					}
					old, next := makeConfig(oldEdge), makeConfig(newEdge)
					result := next.Diff(old, ComponentResourceID("source_"+name))
					consumer := result.FindResource(ComponentResourceID("consumer"))
					require.Equal(t, oldEdge != newEdge, consumer.DirectSummary().HasChanged)
					require.Equal(t, oldEdge && newEdge, consumer.Impacted, "a directly edited consumer is not impact-only")
					require.Equal(t, oldEdge || newEdge, result.FindResource(ComponentResourceID("downstream")).Impacted, "removed edges still invalidate surviving consumers")
					require.False(t, result.FindResource(ComponentResourceID("isolated")).Summary().HasChanged)
				}
			}
		}
	})
}

func TestAppConfigDiff_OperationRoleExtractionProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		name := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		for _, principal := range []string{"nuon::component:*", "nuon::component:api", "nuon::action:*", "nuon::action:inspect", "nuon::sandbox"} {
			operation := OperationDeploy
			if strings.HasPrefix(principal, "nuon::action:") {
				operation = OperationTrigger
			} else if principal == "nuon::sandbox" {
				operation = OperationProvision
			}
			cfg := &AppConfig{
				Sandbox:        &AppSandboxConfig{},
				Components:     ComponentList{{Name: "api", Type: JobComponentType, Job: &JobComponentConfig{}}, {Name: "worker", Type: JobComponentType, Job: &JobComponentConfig{}}},
				Actions:        []*ActionConfig{{Name: "inspect", Steps: []*ActionStepConfig{{Name: "inspect", Command: "echo ready"}}}, {Name: "cleanup", Steps: []*ActionStepConfig{{Name: "cleanup", Command: "echo ready"}}}},
				OperationRoles: &OperationRolesConfig{Type: OperationRuleConfigTypeMatrix, RuleMatrix: []*OperationRoleRule{{Principal: principal, Operation: operation, RoleName: name}}},
			}
			require.NoError(t, cfg.OperationRoles.Validate())
			result := cfg.Diff(cfg, RoleResourceID(name))
			expected := map[configdiff.NodeID]bool{
				ComponentResourceID("api"):    principal == "nuon::component:*" || principal == "nuon::component:api",
				ComponentResourceID("worker"): principal == "nuon::component:*",
				ActionResourceID("inspect"):   principal == "nuon::action:*" || principal == "nuon::action:inspect",
				ActionResourceID("cleanup"):   principal == "nuon::action:*",
				SandboxResourceID:             principal == "nuon::sandbox",
			}
			for id, impacted := range expected {
				node := result.FindResource(id)
				require.NotNil(t, node)
				require.False(t, node.DirectSummary().HasChanged)
				require.Equal(t, impacted, node.Impacted, "%s via %s", id, principal)
				if impacted {
					require.Equal(t, []configdiff.ImpactReason{{From: RoleResourceID(name), Edge: configdiff.EdgeReasonOperationRole}}, node.ImpactReasons)
				}
			}
		}
	})
}

func TestAppConfigDiff_InfrastructureReferenceProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		name := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		for _, source := range []struct {
			expression string
			id         configdiff.NodeID
			reason     configdiff.EdgeReason
		}{
			{".nuon.inputs.inputs." + name, InputResourceID(name), configdiff.EdgeReasonInputReference},
			{".nuon.install.inputs." + name, InputResourceID(name), configdiff.EdgeReasonInputReference},
			{".nuon.secrets." + name, SecretResourceID(name), configdiff.EdgeReasonSecretReference},
			{".nuon.install_stack.outputs.region", StackResourceID, configdiff.EdgeReasonInstallStackOutput},
			{".nuon.install.name", RoleResourceID(name), configdiff.EdgeReasonOperationRole},
		} {
			cfg := &AppConfig{
				Sandbox: &AppSandboxConfig{
					VarsMap:        map[string]string{"region": "{{ " + source.expression + " }}"},
					OperationRoles: []EntityOperationRole{{Operation: OperationProvision, RoleName: name}},
				},
				Components: ComponentList{
					{Name: "consumer", Type: JobComponentType, Job: &JobComponentConfig{EnvVarMap: map[string]string{"REGION": "{{ .nuon.sandbox.outputs.region }}"}}},
					{Name: "isolated", Type: JobComponentType, Job: &JobComponentConfig{}},
				},
			}
			result := cfg.Diff(cfg, source.id)
			sandbox := result.FindResource(SandboxResourceID)
			require.False(t, sandbox.DirectSummary().HasChanged)
			require.True(t, sandbox.Impacted, "sandbox must consume %s", source.expression)
			require.Contains(t, sandbox.ImpactReasons, configdiff.ImpactReason{From: source.id, Edge: source.reason})
			require.True(t, result.FindResource(ComponentResourceID("consumer")).Impacted)
			require.False(t, result.FindResource(ComponentResourceID("isolated")).Summary().HasChanged)
		}
		cfg := &AppConfig{Stack: &StackConfig{CustomNestedStacks: []CustomNestedStack{{
			Name: "acme", TemplateURL: "https://example.com/acme.yaml", Parameters: map[string]string{"Region": "{{ .nuon.install.inputs." + name + " }}"},
		}}}}
		result := cfg.Diff(cfg, InputResourceID(name))
		stack := result.FindResource(StackResourceID)
		require.False(t, stack.DirectSummary().HasChanged)
		require.True(t, stack.Impacted)
		require.Equal(t, []configdiff.ImpactReason{{From: InputResourceID(name), Edge: configdiff.EdgeReasonInputReference}}, stack.ImpactReasons)
	})
}

func appConfigWithNamedPolicy(named []NamedIAMPolicy) *AppConfig {
	cfg := appConfigWithRoleDependency(`{"Statement":[{"Action":"s3:GetObject"}]}`)
	cfg.Permissions.NamedPolicies = named
	return cfg
}

func appConfigWithRoleDependency(policyContents string) *AppConfig {
	return &AppConfig{
		Runner: &AppRunnerConfig{},
		Stack:  &StackConfig{},
		Permissions: &PermissionsConfig{
			MaintenanceRole: &AppAWSIAMRole{
				Name: "maintenance",
				Policies: []AppAWSIAMPolicy{{
					Name:     "maintenance",
					Contents: policyContents,
				}},
			},
		},
		Components: ComponentList{{
			Name: "api",
			Type: HelmChartComponentType,
			OperationRoles: []EntityOperationRole{{
				Operation: OperationDeploy,
				RoleName:  "maintenance",
			}},
			HelmChart: &HelmChartComponentConfig{},
		}},
	}
}

func findResourceDiff(root *configdiff.Diff, id configdiff.NodeID) *configdiff.Diff {
	if root == nil {
		return nil
	}
	if root.ResourceID == id {
		return root
	}
	for _, child := range root.Children {
		if result := findResourceDiff(child, id); result != nil {
			return result
		}
	}
	return nil
}
