package build

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
)

func TestValidateAzureBuiltInRoles(t *testing.T) {
	t.Run("mapped name is accepted", func(t *testing.T) {
		err := ValidateAzureBuiltInRoles("install-preview-operations", []config.AppAWSIAMPolicy{
			{Name: "k8s", AzureBuiltInRoles: []string{"Azure Kubernetes Service RBAC Reader"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("literal GUID is accepted", func(t *testing.T) {
		err := ValidateAzureBuiltInRoles("install-preview-operations", []config.AppAWSIAMPolicy{
			{Name: "k8s", AzureBuiltInRoles: []string{"7f6c6a51-bcf8-42ba-9220-52d62157d7db"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Without this the value is forwarded to ARM verbatim and the customer's
	// stack deploy fails with InvalidRoleDefinitionId.
	t.Run("unknown name is rejected at sync", func(t *testing.T) {
		err := ValidateAzureBuiltInRoles("install-preview-operations", []config.AppAWSIAMPolicy{
			{Name: "k8s", AzureBuiltInRoles: []string{"Azure Kubernetes Service RBAC Readr"}},
		})
		if err == nil {
			t.Fatal("expected an error for an unmapped role name")
		}
		for _, want := range []string{"install-preview-operations", "k8s", "Azure Kubernetes Service RBAC Readr"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error should mention %q, got: %v", want, err)
			}
		}
		if !strings.Contains(err.Error(), "Azure Kubernetes Service RBAC Reader") {
			t.Errorf("error should list valid names, got: %v", err)
		}
	})

	t.Run("policies without azure roles are untouched", func(t *testing.T) {
		err := ValidateAzureBuiltInRoles("install-provision", []config.AppAWSIAMPolicy{
			{Name: "aws", Contents: `{"Statement":[]}`},
			{Name: "azure-actions", AzureActions: []string{"Microsoft.Network/dnsZones/read"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
