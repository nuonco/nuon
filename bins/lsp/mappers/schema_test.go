package mappers

import (
	"slices"
	"testing"

	"github.com/nuonco/nuon/pkg/config/schema"
)

func TestBuildPropertyMap_BranchModeEnums(t *testing.T) {
	branchSchema, err := schema.LookupSchemaType("branch")
	if err != nil {
		t.Fatalf("failed to load branch schema: %v", err)
	}

	hierarchicalMap, _ := BuildPropertyMap(branchSchema)
	tests := map[string][]any{
		"run":     {"push", "on_tag", "on_github_label", "manual_only"},
		"preview": {"plan-only", "apply", "build-only"},
	}
	for table, expected := range tests {
		mode, ok := hierarchicalMap[table]["mode"]
		if !ok {
			t.Fatalf("%s.mode not found", table)
		}
		for _, value := range expected {
			if !slices.Contains(mode.Enum, value) {
				t.Errorf("%s.mode enum missing %q: %v", table, value, mode.Enum)
			}
		}
	}
}

func TestBuildPropertyMap(t *testing.T) {
	// Use the real helm schema from the project
	helmSchema, err := schema.LookupSchemaType("helm")
	if err != nil {
		t.Fatalf("failed to load helm schema: %v", err)
	}
	if helmSchema == nil {
		t.Fatal("helm schema is nil")
	}

	// Build the hierarchical property map
	hierarchicalMap, _ := BuildPropertyMap(helmSchema)

	// Verify the map is not empty
	if len(hierarchicalMap) == 0 {
		t.Error("hierarchical map should not be empty for helm schema")
	}

	// Test root-level properties (table path "")
	t.Run("root level properties", func(t *testing.T) {
		rootProps, ok := hierarchicalMap[""]
		if !ok {
			t.Fatal("root level properties not found")
		}

		expectedKeys := []string{
			"chart_name",
			"namespace",
			"values_file",
			"public_repo",
			"connected_repo",
			"helm_repo",
		}

		for _, key := range expectedKeys {
			if prop, ok := rootProps[key]; !ok {
				t.Errorf("expected root property %q not found", key)
			} else if prop == nil {
				t.Errorf("root property %q is nil", key)
			}
		}
	})

	// Test nested properties are in their correct table paths (resolved from $ref)
	t.Run("public_repo nested properties", func(t *testing.T) {
		publicRepoProps, ok := hierarchicalMap["public_repo"]
		if !ok {
			t.Fatal("public_repo table not found")
		}

		expectedKeys := []string{"repo", "directory"}
		for _, key := range expectedKeys {
			if prop, ok := publicRepoProps[key]; !ok {
				t.Errorf("expected property %q not found in public_repo", key)
			} else if prop == nil {
				t.Errorf("property %q in public_repo is nil", key)
			}
		}
	})

	t.Run("connected_repo nested properties", func(t *testing.T) {
		connectedRepoProps, ok := hierarchicalMap["connected_repo"]
		if !ok {
			t.Fatal("connected_repo table not found")
		}

		expectedKeys := []string{"repo", "directory", "branch"}
		for _, key := range expectedKeys {
			if prop, ok := connectedRepoProps[key]; !ok {
				t.Errorf("expected property %q not found in connected_repo", key)
			} else if prop == nil {
				t.Errorf("property %q in connected_repo is nil", key)
			}
		}
	})

	t.Run("values_file nested properties", func(t *testing.T) {
		valuesFileProps, ok := hierarchicalMap["values_file"]
		if !ok {
			t.Fatal("values_file table not found")
		}

		expectedKeys := []string{"source", "path"}
		for _, key := range expectedKeys {
			if prop, ok := valuesFileProps[key]; !ok {
				t.Errorf("expected property %q not found in values_file", key)
			} else if prop == nil {
				t.Errorf("property %q in values_file is nil", key)
			}
		}
	})

	// Verify specific property details
	t.Run("chart_name property", func(t *testing.T) {
		rootProps, ok := hierarchicalMap[""]
		if !ok {
			t.Fatal("root level not found")
		}
		prop, ok := rootProps["chart_name"]
		if !ok {
			t.Fatal("chart_name not found")
		}
		if prop.Type != "string" {
			t.Errorf("chart_name type = %q, want %q", prop.Type, "string")
		}
	})

	t.Run("namespace property", func(t *testing.T) {
		rootProps, ok := hierarchicalMap[""]
		if !ok {
			t.Fatal("root level not found")
		}
		prop, ok := rootProps["namespace"]
		if !ok {
			t.Fatal("namespace not found")
		}
		if prop.Type != "string" {
			t.Errorf("namespace type = %q, want %q", prop.Type, "string")
		}
	})
}

func TestBuildPropertyMapNilSchema(t *testing.T) {
	hierarchicalMap, _ := BuildPropertyMap(nil)
	if len(hierarchicalMap) != 0 {
		t.Errorf("expected empty map for nil schema, got %d entries", len(hierarchicalMap))
	}
}
