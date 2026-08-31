package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
)

func TestValidateTemplateRefs_MissingComponent(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Components: []*config.Component{{
			Name: "consumer",
			Type: config.TerraformModuleComponentType,
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "ghost",
				Input: "nuon.components.ghost.outputs.id",
			}},
		}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown component \"ghost\"")
	assert.Contains(t, err.Error(), "component:consumer")
}

func TestValidateTemplateRefs_MissingAction(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Actions: []*config.ActionConfig{{
			Name: "runner",
			References: []refs.Ref{{
				Type:  refs.RefTypeActions,
				Name:  "ghost_action",
				Input: "nuon.actions.ghost_action.outputs.token",
			}},
		}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action \"ghost_action\"")
}

func TestValidateTemplateRefs_DeclaredComponentAndAction(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Components: []*config.Component{{
			Name: "database",
			Type: config.TerraformModuleComponentType,
		}},
		Actions: []*config.ActionConfig{{
			Name: "healthcheck",
		}, {
			Name: "consumer",
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "database",
				Input: "nuon.components.database.outputs.host",
			}, {
				Type:  refs.RefTypeActions,
				Name:  "healthcheck",
				Input: "nuon.actions.healthcheck.outputs.status",
			}},
		}},
	})
	require.NoError(t, err)
}

func TestValidateTemplateRefs_SandboxRejectsComponentAndAction(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Components: []*config.Component{{Name: "nlb"}},
		Actions:    []*config.ActionConfig{{Name: "setup"}},
		Sandbox: &config.AppSandboxConfig{
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "nlb",
				Input: "nuon.components.nlb.outputs.dns_name",
			}, {
				Type:  refs.RefTypeActions,
				Name:  "setup",
				Input: "nuon.actions.setup.outputs.token",
			}},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sandbox references nuon.components.nlb.outputs.dns_name")
	assert.Contains(t, err.Error(), "sandbox references nuon.actions.setup.outputs.token")
	assert.Contains(t, err.Error(), "not available when rendering sandbox config")
}

func TestValidateTemplateRefs_StackRejectsLateBoundRefs(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Components: []*config.Component{{Name: "nlb"}},
		Actions:    []*config.ActionConfig{{Name: "setup"}},
		Stack: &config.StackConfig{
			Name:                    "my-stack-{{ .nuon.components.nlb.outputs.dns_name }}",
			Description:             "zone {{ .nuon.sandbox.outputs.cluster }}",
			VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/{{ .nuon.actions.setup.outputs.token }}.yaml",
			RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/{{ .nuon.install_stack.outputs.region }}.yaml",
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name references nuon.components.nlb.outputs.dns_name")
	assert.Contains(t, err.Error(), "description references nuon.sandbox.outputs.cluster")
	assert.Contains(t, err.Error(), "vpc_nested_template_url references nuon.actions.setup.outputs.token")
	assert.Contains(t, err.Error(), "runner_nested_template_url references nuon.install_stack.outputs.region")
	assert.Contains(t, err.Error(), "not populated when the install stack is generated")
}

func TestValidateTemplateRefs_CustomNestedStackParameter(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Stack: &config.StackConfig{
			CustomNestedStacks: []config.CustomNestedStack{{
				Name:        "k8s",
				TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml",
				Parameters: map[string]string{
					"Cluster": "{{ .nuon.sandbox.outputs.cluster }}",
				},
			}},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "custom_nested_stacks[0].parameters.Cluster")
	assert.Contains(t, err.Error(), "not populated when the install stack is generated")
}

func TestValidateTemplateRefs_AllowedInstallAppOrgAndStackRefs(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Components: []*config.Component{{
			Name: "api",
			References: []refs.Ref{{
				Type:  refs.RefTypeInstallInputs,
				Name:  "cluster_name",
				Input: "nuon.install.inputs.cluster_name",
			}, {
				Type:  refs.RefTypeInputs,
				Name:  "version",
				Input: "nuon.inputs.inputs.version",
			}},
		}},
		Stack: &config.StackConfig{
			Name:        "stack-{{ .nuon.install.id }}",
			Description: "app {{ .nuon.app.name }} org {{ .nuon.org.id }}",
			CustomNestedStacks: []config.CustomNestedStack{{
				Name:        "k8s",
				TemplateURL: "https://s3.amazonaws.com/bucket/{{ .nuon.install.id }}.yaml",
				Parameters: map[string]string{
					"Namespaces": "{{ .nuon.install.inputs.namespaces }}",
					"Inputs":     "{{ .nuon.inputs.inputs.cluster_version }}",
					"InstallID":  "vendor-{{ .nuon.install.id }}-service",
				},
			}},
		},
		Sandbox: &config.AppSandboxConfig{
			References: []refs.Ref{{
				Type:  refs.RefTypeInstallStack,
				Name:  "region",
				Input: "nuon.install_stack.outputs.region",
			}, {
				Type:  refs.RefTypeInstallInputs,
				Name:  "cluster_name",
				Input: "nuon.install.inputs.cluster_name",
			}},
		},
	})
	require.NoError(t, err)
}

func TestValidateTemplateRefs_CollectsAllFindings(t *testing.T) {
	err := ValidateTemplateRefs(&config.AppConfig{
		Components: []*config.Component{{
			Name: "consumer",
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "ghost",
				Input: "nuon.components.ghost.outputs.id",
			}},
		}},
		Sandbox: &config.AppSandboxConfig{
			References: []refs.Ref{{
				Type:  refs.RefTypeComponents,
				Name:  "ghost",
				Input: "nuon.components.ghost.outputs.id",
			}},
		},
		Stack: &config.StackConfig{
			Name: "stack-{{ .nuon.components.ghost.outputs.id }}",
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown component \"ghost\"")
	assert.Contains(t, err.Error(), "sandbox references")
	assert.Contains(t, err.Error(), "name references nuon.components.ghost.outputs.id")
	assert.Contains(t, err.Error(), "invalid template references:")
}
