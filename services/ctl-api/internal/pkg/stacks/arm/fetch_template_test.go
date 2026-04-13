package arm

import (
	"encoding/json"
	"testing"
)

func TestValidateARMTemplate_Valid(t *testing.T) {
	tmpl := &armTemplateShape{
		Resources: []armTemplateResource{
			{Type: "Microsoft.ManagedIdentity/userAssignedIdentities", Name: "myIdentity"},
			{
				Type:      "Microsoft.Resources/deploymentScripts",
				Name:      "enableApis",
				DependsOn: json.RawMessage(`["myIdentity"]`),
			},
		},
	}

	if err := validateARMTemplate(tmpl); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateARMTemplate_SubscriptionLevelDeployment(t *testing.T) {
	tmpl := &armTemplateShape{
		Resources: []armTemplateResource{
			{
				Type:           "Microsoft.Resources/deployments",
				Name:           "roleAssignment",
				SubscriptionId: "[subscription().subscriptionId]",
			},
		},
	}

	err := validateARMTemplate(tmpl)
	if err == nil {
		t.Fatal("expected validation error for subscription-level nested deployment")
	}

	if got := err.Error(); !contains(got, "subscription-level nested deployments") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestValidateARMTemplate_InvalidDependsOn(t *testing.T) {
	tmpl := &armTemplateShape{
		Resources: []armTemplateResource{
			{Type: "Microsoft.Storage/storageAccounts", Name: "storage"},
			{
				Type:      "Microsoft.Resources/deploymentScripts",
				Name:      "script",
				DependsOn: json.RawMessage(`["nonExistentResource"]`),
			},
		},
	}

	err := validateARMTemplate(tmpl)
	if err == nil {
		t.Fatal("expected validation error for invalid dependsOn reference")
	}

	if got := err.Error(); !contains(got, "nonExistentResource") || !contains(got, "not defined") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestValidateARMTemplate_DependsOnExpressionSkipped(t *testing.T) {
	tmpl := &armTemplateShape{
		Resources: []armTemplateResource{
			{
				Type:      "Microsoft.Resources/deploymentScripts",
				Name:      "script",
				DependsOn: json.RawMessage(`["[resourceId('Microsoft.Network/virtualNetworks', 'myVnet')]"]`),
			},
		},
	}

	if err := validateARMTemplate(tmpl); err != nil {
		t.Fatalf("ARM expressions in dependsOn should be skipped, got: %v", err)
	}
}

func TestValidateARMTemplate_InlineNestedSubscriptionDeployment(t *testing.T) {
	tmpl := &armTemplateShape{
		Resources: []armTemplateResource{
			{
				Type: "Microsoft.Resources/deployments",
				Name: "outerDeployment",
				Properties: &armResourceProperties{
					Template: &struct {
						Resources []armTemplateResource `json:"resources,omitempty"`
					}{
						Resources: []armTemplateResource{
							{
								Type:           "Microsoft.Resources/deployments",
								Name:           "innerSubscriptionDeployment",
								SubscriptionId: "[subscription().subscriptionId]",
							},
						},
					},
				},
			},
		},
	}

	err := validateARMTemplate(tmpl)
	if err == nil {
		t.Fatal("expected validation error for subscription-level deployment inside inline template")
	}

	if got := err.Error(); !contains(got, "inline template") || !contains(got, "innerSubscriptionDeployment") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestValidateARMTemplate_MultipleErrors(t *testing.T) {
	tmpl := &armTemplateShape{
		Resources: []armTemplateResource{
			{
				Type:           "Microsoft.Resources/deployments",
				Name:           "badDeployment",
				SubscriptionId: "[subscription().subscriptionId]",
				DependsOn:      json.RawMessage(`["ghost"]`),
			},
		},
	}

	err := validateARMTemplate(tmpl)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	got := err.Error()
	if !contains(got, "subscription-level") {
		t.Errorf("missing subscription-level error: %s", got)
	}
	if !contains(got, "ghost") {
		t.Errorf("missing dependsOn error: %s", got)
	}
}

func TestValidateARMTemplate_NoDependsOn(t *testing.T) {
	tmpl := &armTemplateShape{
		Resources: []armTemplateResource{
			{Type: "Microsoft.Storage/storageAccounts", Name: "storage"},
		},
	}

	if err := validateARMTemplate(tmpl); err != nil {
		t.Fatalf("expected no error for resource without dependsOn, got: %v", err)
	}
}

func TestValidateARMTemplate_EmptyTemplate(t *testing.T) {
	tmpl := &armTemplateShape{}

	if err := validateARMTemplate(tmpl); err != nil {
		t.Fatalf("expected no error for empty template, got: %v", err)
	}
}

func TestParseDependsOn_Array(t *testing.T) {
	deps, err := parseDependsOn(json.RawMessage(`["a", "b"]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 2 || deps[0] != "a" || deps[1] != "b" {
		t.Errorf("expected [a b], got %v", deps)
	}
}

func TestParseDependsOn_String(t *testing.T) {
	deps, err := parseDependsOn(json.RawMessage(`"singleDep"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 || deps[0] != "singleDep" {
		t.Errorf("expected [singleDep], got %v", deps)
	}
}

func TestParseDependsOn_Nil(t *testing.T) {
	deps, err := parseDependsOn(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected empty, got %v", deps)
	}
}

func TestParseDependsOn_Invalid(t *testing.T) {
	_, err := parseDependsOn(json.RawMessage(`123`))
	if err == nil {
		t.Fatal("expected error for non-string/non-array value")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
