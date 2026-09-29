package dir

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestSourceFileSetter_InterfaceAssertion(t *testing.T) {
	policy := &config.AppPolicy{}

	setter, ok := interface{}(policy).(sourceFileSetter)
	require.True(t, ok, "*AppPolicy should implement sourceFileSetter")

	setter.SetSourceFile("/path/to/policy.rego")
	require.Equal(t, "/path/to/policy.rego", policy.SourceFile)
}

func TestNameFromSourceFileSetter_InterfaceAssertion(t *testing.T) {
	policy := &config.AppPolicy{
		SourceFile: "/app/policies/block-mutable-tags.rego",
	}

	setter, ok := interface{}(policy).(nameFromSourceFileSetter)
	require.True(t, ok, "*AppPolicy should implement nameFromSourceFileSetter")

	setter.SetNameFromSourceFile()
	require.Equal(t, "block-mutable-tags", policy.Name)
}

func TestValueType_DoesNotImplementInterfaces(t *testing.T) {
	policy := config.AppPolicy{}

	_, ok := interface{}(policy).(sourceFileSetter)
	require.False(t, ok, "AppPolicy (value type) should NOT implement sourceFileSetter - methods have pointer receivers")

	_, ok = interface{}(policy).(nameFromSourceFileSetter)
	require.False(t, ok, "AppPolicy (value type) should NOT implement nameFromSourceFileSetter - methods have pointer receivers")
}

func TestSkipPermissionsPoliciesAsRoles(t *testing.T) {
	require.True(t, skipPermissionsPoliciesAsRoles("permissions", "permissions/policies/logs.toml"))
	require.False(t, skipPermissionsPoliciesAsRoles("permissions", "permissions/provision.toml"))
	require.False(t, skipPermissionsPoliciesAsRoles("permissions/policies", "permissions/policies/logs.toml"))
}
