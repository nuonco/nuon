package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeOpentofuModuleComponent(t *testing.T) {
	out, err := decodeComponent(
		reflect.TypeOf(map[string]interface{}{}),
		reflect.TypeOf(Component{}),
		map[string]interface{}{
			"type":              string(OpentofuModuleComponentType),
			"name":              "acme-infra",
			"open_tofu_version": "1.9.0",
			"vars":              map[string]interface{}{"region": "us-east-1"},
			"env_vars":          map[string]interface{}{"TF_LOG": "info"},
		},
		parseOptions(),
	)
	require.NoError(t, err)

	comp := out.(Component)
	require.Equal(t, OpentofuModuleComponentType, comp.Type)
	require.NotNil(t, comp.OpentofuModule)
	assert.Equal(t, "1.9.0", comp.OpentofuModule.OpenTofuVersion)
	assert.Equal(t, map[string]string{"region": "us-east-1"}, comp.OpentofuModule.VarsMap)
	assert.Equal(t, map[string]string{"TF_LOG": "info"}, comp.OpentofuModule.EnvVarMap)

	require.NoError(t, comp.Validate())
}

func TestOpentofuModuleComponentValidateTeardown(t *testing.T) {
	comp := &Component{
		Type: OpentofuModuleComponentType,
		Name: "acme-infra",
		OpentofuModule: &OpentofuModuleComponentConfig{
			OpenTofuVersion: "1.9.0",
			Teardown:        &ComponentTeardownConfig{Retain: []string{"aws_s3_bucket.data", "aws_s3_bucket.data"}},
		},
	}
	err := comp.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate address")
}

func TestOpentofuModuleComponentOverrideInputs(t *testing.T) {
	inputs := SyntheticComponentOverrideInputs(ComponentList{
		{Name: "acme-infra", Type: OpentofuModuleComponentType},
	})
	require.Len(t, inputs, 1)
	assert.Equal(t, ComponentOverrideKindTFVars, inputs[0].Kind)
	// OpenTofu reuses the terraform tf_vars override input name so
	// install-level overrides work uniformly.
	assert.Equal(t, TFVarsOverrideInputName("acme-infra"), inputs[0].Name)
}

func TestOpentofuModuleComponentAPIType(t *testing.T) {
	assert.Equal(t, "opentofu_module", string(OpentofuModuleComponentType.APIType()))
}

func TestOpentofuModulePolicyType(t *testing.T) {
	assert.Equal(t, AppPolicyTypeOpentofuModule, policyType(OpentofuModuleComponentType))
}
