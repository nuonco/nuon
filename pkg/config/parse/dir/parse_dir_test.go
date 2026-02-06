package dir

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestSourceFileSetter_InterfaceAssertion(t *testing.T) {
	// This test verifies that *AppPolicy correctly implements sourceFileSetter
	// The bug was using objValue.Interface() (the dereferenced value) instead of obj (the pointer)

	policy := &config.AppPolicy{}

	// This is what the fixed code does - assert on the pointer type
	setter, ok := interface{}(policy).(sourceFileSetter)
	require.True(t, ok, "*AppPolicy should implement sourceFileSetter")

	setter.SetSourceFile("/path/to/policy.rego")
	require.Equal(t, "/path/to/policy.rego", policy.SourceFile)
}

func TestNameFromSourceFileSetter_InterfaceAssertion(t *testing.T) {
	// This test verifies that *AppPolicy correctly implements nameFromSourceFileSetter

	policy := &config.AppPolicy{
		SourceFile: "/app/policies/block-mutable-tags.rego",
	}

	// This is what the fixed code does - assert on the pointer type
	setter, ok := interface{}(policy).(nameFromSourceFileSetter)
	require.True(t, ok, "*AppPolicy should implement nameFromSourceFileSetter")

	setter.SetNameFromSourceFile()
	require.Equal(t, "block-mutable-tags", policy.Name)
}

func TestValueType_DoesNotImplementInterfaces(t *testing.T) {
	// This test demonstrates the bug - value types don't implement pointer receiver interfaces

	policy := config.AppPolicy{}

	// Value type does NOT implement the interface (methods have pointer receivers)
	_, ok := interface{}(policy).(sourceFileSetter)
	require.False(t, ok, "AppPolicy (value type) should NOT implement sourceFileSetter - methods have pointer receivers")

	_, ok = interface{}(policy).(nameFromSourceFileSetter)
	require.False(t, ok, "AppPolicy (value type) should NOT implement nameFromSourceFileSetter - methods have pointer receivers")
}
