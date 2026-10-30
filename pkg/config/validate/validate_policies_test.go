package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestValidateOPAPolicy(t *testing.T) {
	t.Run("accepts deny rule", func(t *testing.T) {
		require.NoError(t, ValidateOPAPolicy("package nuon\ndeny := []"))
	})

	t.Run("accepts warn rule", func(t *testing.T) {
		require.NoError(t, ValidateOPAPolicy("package nuon\nwarn := []"))
	})

	t.Run("accepts named checks with the Nuon library", func(t *testing.T) {
		require.NoError(t, ValidateOPAPolicy(`package nuon.checks
import data.nuon.lib.check
check_example contains check.finding("review")`))
	})

	t.Run("rejects another package", func(t *testing.T) {
		err := ValidateOPAPolicy("package other\ndeny := []")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must use package nuon")
	})

	t.Run("rejects policy without decisions", func(t *testing.T) {
		err := ValidateOPAPolicy("package nuon\nallow := true")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must define at least one deny or warn rule")
	})
}

func TestValidatePolicyType(t *testing.T) {
	tests := map[string]struct {
		input    config.AppPolicyType
		expected bool
	}{
		"kubernetes_cluster":  {config.AppPolicyTypeKubernetesCluster, false},
		"terraform_module":    {config.AppPolicyTypeTerraformModule, false},
		"helm_chart":          {config.AppPolicyTypeHelmChart, false},
		"kubernetes_manifest": {config.AppPolicyTypeKubernetesManifest, false},
		"docker_build":        {config.AppPolicyTypeDockerBuild, false},
		"container_image":     {config.AppPolicyTypeContainerImage, false},
		"sandbox":             {config.AppPolicyTypeSandbox, false},
		"invalid_policy_type": {config.AppPolicyType("invalid_policy_type"), true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePolicyType(test.input)
			assert.Equal(t, (err != nil), test.expected, "Expected error for policy type %s: %v, got: %v", test.input, test.expected, err)
		})
	}
}

func TestSkipApplyPolicyScope(t *testing.T) {
	contents := `package nuon.checks
import data.nuon.lib.check
# METADATA
# custom:
#   action: skip_apply
check_defer contains check.finding("defer replacement") if { input.replace }
`
	for _, policyType := range []config.AppPolicyType{
		config.AppPolicyTypeTerraformModule, config.AppPolicyTypeSandbox,
		config.AppPolicyTypeHelmChart, config.AppPolicyTypeContainerImage,
	} {
		for _, teardown := range []*bool{nil, new(true), new(false)} {
			cfg := &config.AppConfig{Policies: &config.PoliciesConfig{Policies: []config.AppPolicy{{
				Type: policyType, Engine: config.AppPolicyEngineOPA,
				Contents: contents, ApplyOnTeardown: teardown,
			}}}}
			err := ValidatePolicies(cfg)
			if policyType == config.AppPolicyTypeTerraformModule || policyType == config.AppPolicyTypeSandbox {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "skip_apply is only supported for terraform_module")
			}
		}
	}
}

func TestValidatePolicyEngine(t *testing.T) {
	tests := map[string]struct {
		input    config.AppPolicyEngine
		expected bool
	}{
		"kyverno":        {config.AppPolicyEngineKyverno, false},
		"opa":            {config.AppPolicyEngineOPA, false},
		"empty_allowed":  {"", false}, // empty is allowed for backwards compatibility
		"invalid_engine": {config.AppPolicyEngine("invalid_engine"), true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePolicyEngine(test.input)
			assert.Equal(t, (err != nil), test.expected, "Expected error for engine %s: %v, got: %v", test.input, test.expected, err)
		})
	}
}

func TestValidatePolicyTypeEngineCompatibility(t *testing.T) {
	tests := map[string]struct {
		policyType config.AppPolicyType
		engine     config.AppPolicyEngine
		expected   bool
	}{
		// kubernetes_cluster only supports kyverno
		"kubernetes_cluster_kyverno":   {config.AppPolicyTypeKubernetesCluster, config.AppPolicyEngineKyverno, false},
		"kubernetes_cluster_opa_error": {config.AppPolicyTypeKubernetesCluster, config.AppPolicyEngineOPA, true},
		// component-based types only support OPA
		"terraform_module_kyverno_error":    {config.AppPolicyTypeTerraformModule, config.AppPolicyEngineKyverno, true},
		"terraform_module_opa":              {config.AppPolicyTypeTerraformModule, config.AppPolicyEngineOPA, false},
		"helm_chart_kyverno_error":          {config.AppPolicyTypeHelmChart, config.AppPolicyEngineKyverno, true},
		"helm_chart_opa":                    {config.AppPolicyTypeHelmChart, config.AppPolicyEngineOPA, false},
		"kubernetes_manifest_kyverno_error": {config.AppPolicyTypeKubernetesManifest, config.AppPolicyEngineKyverno, true},
		"kubernetes_manifest_opa":           {config.AppPolicyTypeKubernetesManifest, config.AppPolicyEngineOPA, false},
		// empty engine skips check
		"kubernetes_cluster_empty_engine": {config.AppPolicyTypeKubernetesCluster, "", false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePolicyTypeEngineCompatibility(test.policyType, test.engine)
			assert.Equal(t, (err != nil), test.expected, "Expected error: %v, got: %v", test.expected, err)
		})
	}
}

func TestValidatePoliciesUniqueNames(t *testing.T) {
	policy := func(name string, typ config.AppPolicyType, components []string) config.AppPolicy {
		return config.AppPolicy{
			Name:       name,
			Type:       typ,
			Engine:     config.AppPolicyEngineOPA,
			Contents:   "package nuon\ndeny := []",
			Components: components,
		}
	}

	t.Run("accepts distinct names", func(t *testing.T) {
		cfg := &config.AppConfig{Policies: &config.PoliciesConfig{Policies: []config.AppPolicy{
			policy("sandbox-ports", config.AppPolicyTypeSandbox, nil),
			policy("module-ports", config.AppPolicyTypeTerraformModule, []string{"*"}),
		}}}
		require.NoError(t, ValidatePolicyNames(cfg))
	})

	t.Run("rejects duplicate names", func(t *testing.T) {
		cfg := &config.AppConfig{Policies: &config.PoliciesConfig{Policies: []config.AppPolicy{
			policy("warn-public-sensitive-ports", config.AppPolicyTypeSandbox, nil),
			policy("warn-public-sensitive-ports", config.AppPolicyTypeTerraformModule, []string{"*"}),
		}}}
		err := ValidatePolicyNames(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `duplicate policy name "warn-public-sensitive-ports"`)
	})

	t.Run("accepts a nil config", func(t *testing.T) {
		require.NoError(t, ValidatePolicyNames(nil))
	})

	t.Run("accepts missing policies", func(t *testing.T) {
		require.NoError(t, ValidatePolicyNames(&config.AppConfig{}))
		require.NoError(t, ValidateUniquePolicyNames(nil))
	})

	t.Run("accepts one unnamed policy", func(t *testing.T) {
		require.NoError(t, ValidateUniquePolicyNames([]config.AppPolicy{{}}))
	})

	t.Run("rejects two unnamed policies", func(t *testing.T) {
		err := ValidateUniquePolicyNames([]config.AppPolicy{{}, {}})
		require.Error(t, err)
		assert.EqualError(t, err, "policies must have unique names")
	})

	t.Run("keeps an explicit name over the contents path", func(t *testing.T) {
		cfg := &config.AppConfig{Policies: &config.PoliciesConfig{Policies: []config.AppPolicy{
			{
				Name:     "sandbox-ports",
				Contents: "./policies/warn-public-sensitive-ports.rego",
			},
			{
				Name:     "module-ports",
				Contents: "./policies/warn-public-sensitive-ports.rego",
			},
		}}}
		require.NoError(t, ValidatePolicyNames(cfg))
	})

	t.Run("rejects names derived from the same contents path", func(t *testing.T) {
		cfg := &config.AppConfig{Policies: &config.PoliciesConfig{Policies: []config.AppPolicy{
			{
				Type:     config.AppPolicyTypeSandbox,
				Engine:   config.AppPolicyEngineOPA,
				Contents: "./policies/warn-public-sensitive-ports.rego",
			},
			{
				Type:       config.AppPolicyTypeTerraformModule,
				Engine:     config.AppPolicyEngineOPA,
				Components: []string{"*"},
				Contents:   "./policies/warn-public-sensitive-ports.rego",
			},
		}}}
		err := ValidatePolicyNames(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `duplicate policy name "warn-public-sensitive-ports"`)
	})
}

func TestValidatePolicyComponents(t *testing.T) {
	tests := map[string]struct {
		policyType config.AppPolicyType
		components []string
		expected   bool
	}{
		// an omitted list applies to all components of the type
		"terraform_empty_ok":       {config.AppPolicyTypeTerraformModule, []string{}, false},
		"terraform_nil_ok":         {config.AppPolicyTypeTerraformModule, nil, false},
		"container_image_empty_ok": {config.AppPolicyTypeContainerImage, []string{}, false},
		"sandbox_empty_ok":         {config.AppPolicyTypeSandbox, []string{}, false},
		// populated lists validate the same regardless of type
		"single_component":     {config.AppPolicyTypeTerraformModule, []string{"rds_cluster"}, false},
		"multiple_components":  {config.AppPolicyTypeTerraformModule, []string{"rds_cluster", "vpc"}, false},
		"wildcard_only":        {config.AppPolicyTypeTerraformModule, []string{"*"}, false},
		"wildcard_with_others": {config.AppPolicyTypeTerraformModule, []string{"*", "rds_cluster"}, true},
		"empty_component_name": {config.AppPolicyTypeTerraformModule, []string{"rds_cluster", ""}, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidatePolicyComponents(name, test.policyType, test.components)
			assert.Equal(t, (err != nil), test.expected, "Expected error: %v, got: %v", test.expected, err)
		})
	}
}

func TestValidatePolicyApplyOnTeardown(t *testing.T) {
	no := false

	tests := map[string]struct {
		policyType      config.AppPolicyType
		applyOnTeardown *bool
		expected        bool
	}{
		"terraform_unset_ok":       {config.AppPolicyTypeTerraformModule, nil, false},
		"terraform_false_ok":       {config.AppPolicyTypeTerraformModule, &no, false},
		"helm_false_ok":            {config.AppPolicyTypeHelmChart, &no, false},
		"sandbox_false_ok":         {config.AppPolicyTypeSandbox, &no, false},
		"container_image_unset_ok": {config.AppPolicyTypeContainerImage, nil, false},
		// types never evaluated on teardown
		"container_image_err":    {config.AppPolicyTypeContainerImage, &no, true},
		"docker_build_err":       {config.AppPolicyTypeDockerBuild, &no, true},
		"kubernetes_cluster_err": {config.AppPolicyTypeKubernetesCluster, &no, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidatePolicyApplyOnTeardown(name, test.policyType, test.applyOnTeardown)
			assert.Equal(t, test.expected, err != nil, "Expected error: %v, got: %v", test.expected, err)
		})
	}
}
