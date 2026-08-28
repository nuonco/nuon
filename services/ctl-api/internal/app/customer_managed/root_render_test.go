package customermanaged

import (
	"testing"

	"github.com/stretchr/testify/require"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestVirtualInstallIDMatchesEnvelopeFrozenID(t *testing.T) {
	require.Equal(t, virtualID("vinst", "app-test"), VirtualInstallID("app-test"))
}

func TestRenderConfigForStackCompileRendersRoleTemplates(t *testing.T) {
	cfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{Name: "{{.nuon.install.id}}-provision", DisplayName: "provision role"},
				{Name: "{{ .nuon.install.id }}-maintenance"},
			},
		},
		BreakGlassConfig: app.AppBreakGlassConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{Name: "{{.nuon.install.id}}-break-glass"},
			},
		},
	}

	rendered, err := renderConfigForStackCompile(cfg, "vinst0123456789abcdef", nil)
	require.NoError(t, err)
	require.Equal(t, "vinst0123456789abcdef-provision", rendered.PermissionsConfig.Roles[0].Name)
	require.Equal(t, "vinst0123456789abcdef-maintenance", rendered.PermissionsConfig.Roles[1].Name)
	require.Equal(t, "vinst0123456789abcdef-break-glass", rendered.BreakGlassConfig.Roles[0].Name)

	require.Equal(t, "{{.nuon.install.id}}-provision", cfg.PermissionsConfig.Roles[0].Name, "caller's config must not be mutated")
	require.Equal(t, "{{.nuon.install.id}}-break-glass", cfg.BreakGlassConfig.Roles[0].Name, "caller's config must not be mutated")
}

func TestRenderConfigForStackCompileRendersCloudAccountRegion(t *testing.T) {
	cfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{Name: "maintenance", Policies: []app.AppAWSIAMPolicyConfig{
					{Name: "rds-secrets", Contents: []byte(`{"Resource":"arn:aws:secretsmanager:{{ .nuon.cloud_account.aws.region }}::secret:rds!*"}`)},
				}},
			},
		},
	}

	rendered, err := renderConfigForStackCompile(cfg, "vinst0123456789abcdef", nil)
	require.NoError(t, err)
	require.Contains(t, string(rendered.PermissionsConfig.Roles[0].Policies[0].Contents), "arn:aws:secretsmanager:__NUON_CUSTOMER_MANAGED_STACK_region__::secret:rds!*")
}

func TestRenderConfigForStackCompileSeedsInputPlaceholders(t *testing.T) {
	cfg := &app.AppConfig{
		StackConfig: app.AppStackConfig{
			Name: "stack-{{.nuon.install.inputs.env}}",
		},
	}

	rendered, err := renderConfigForStackCompile(cfg, "vinst0123456789abcdef", []customermanaged.InputSpec{{Name: "env"}})
	require.NoError(t, err)
	require.Equal(t, "stack-"+customermanaged.InputPlaceholder("env"), rendered.StackConfig.Name)
}
