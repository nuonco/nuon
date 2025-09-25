package app

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestOrgBeforeCreateFeatureDefaults(t *testing.T) {
	org := &Org{}

	// Create a mock GORM DB with a context
	db := &gorm.DB{
		Statement: &gorm.Statement{
			Context: context.Background(),
		},
	}

	// Simulate the BeforeCreate hook call
	err := org.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate hook failed: %v", err)
	}

	// Verify all feature flags are present
	if len(org.Features) != 11 {
		t.Errorf("Expected 11 feature flags, got %d", len(org.Features))
	}

	// Verify feature flags that should be disabled by default
	if org.Features["org-dashboard"] != false {
		t.Error("org-dashboard should be disabled by default")
	}
	if org.Features["install-break-glass"] != false {
		t.Error("install-break-glass should be disabled by default")
	}

	// Verify feature flags that should be enabled by default
	expectedEnabled := []string{
		"api-pagination",
		"org-runner",
		"org-settings",
		"org-support",
		"install-delete-components",
		"install-delete",
		"terraform-workspace",
		"dev-command",
		"app-branches",
	}

	for _, feature := range expectedEnabled {
		if org.Features[feature] != true {
			t.Errorf("Feature %s should be enabled by default, got %v", feature, org.Features[feature])
		}
	}
}

func TestOrgBeforeCreatePreservesExistingFeatures(t *testing.T) {
	org := &Org{
		Features: map[string]bool{
			"org-dashboard": true, // Set to true explicitly
		},
	}

	// Create a mock GORM DB with a context
	db := &gorm.DB{
		Statement: &gorm.Statement{
			Context: context.Background(),
		},
	}

	// Simulate the BeforeCreate hook call
	err := org.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate hook failed: %v", err)
	}

	// Verify the explicitly set feature is preserved
	if org.Features["org-dashboard"] != true {
		t.Error("Explicitly set org-dashboard should be preserved")
	}

	// Verify other features get default values
	if org.Features["install-break-glass"] != false {
		t.Error("install-break-glass should get default value (false)")
	}

	if org.Features["api-pagination"] != true {
		t.Error("api-pagination should get default value (true)")
	}
}
