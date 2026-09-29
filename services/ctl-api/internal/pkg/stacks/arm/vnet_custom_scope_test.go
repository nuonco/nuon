package arm

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const rgScopedVNetFixture = `{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "nuonInstallID": {"type": "string"},
    "location":      {"type": "string"},
    "commonTags":    {"type": "object"}
  },
  "resources": [
    {"type": "Microsoft.Network/virtualNetworks", "apiVersion": "2023-04-01", "name": "vnet"}
  ],
  "outputs": {}
}`

func customVNetDeployment(t *testing.T, fixture string) map[string]any {
	t.Helper()

	tmpl := &Templates{cfg: &internal.Config{}}
	inp := vnetInputWithTemplate(t, app.StackDeploymentScopeSubscription, fixture)

	dep, _, _, err := tmpl.getVNetLinkedDeployment(inp, scopeFor(inp))
	if err != nil {
		t.Fatalf("getVNetLinkedDeployment: %v", err)
	}
	return dep
}

func TestVNetLinkedDeployment_RGScopedCustomTemplateTargetsInstallRG(t *testing.T) {
	dep := customVNetDeployment(t, rgScopedVNetFixture)

	if got := dep["resourceGroup"]; got != "[variables('installResourceGroupName')]" {
		t.Errorf("resourceGroup = %v, want the install resource group", got)
	}
	if got, present := dep["location"]; present {
		t.Errorf("resource-group-targeted deployment must not set location, got %v", got)
	}
}

func TestVNetLinkedDeployment_SubscriptionScopedCustomTemplateTargetsSubscription(t *testing.T) {
	dep := customVNetDeployment(t, hoistFixture)

	if got, present := dep["resourceGroup"]; present {
		t.Errorf("subscription-targeted deployment must not set resourceGroup, got %v", got)
	}
	if _, present := dep["location"]; !present {
		t.Error("subscription-targeted deployment requires an explicit location")
	}
}

func TestIsSubscriptionScopedTemplate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema string
		want   bool
	}{
		{"subscription", subscriptionTemplateSchema, true},
		{"resource group", rgTemplateSchema, false},
		{"missing", "", false},
		{"unrecognised", "https://example.com/whatever.json#", false},
		{"no trailing hash", "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSubscriptionScopedTemplate(&armTemplateShape{Schema: tc.schema}); got != tc.want {
				t.Errorf("isSubscriptionScopedTemplate(%q) = %v, want %v", tc.schema, got, tc.want)
			}
		})
	}
}
