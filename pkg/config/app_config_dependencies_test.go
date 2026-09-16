package config

import (
	"testing"

	"github.com/stretchr/testify/require"

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
