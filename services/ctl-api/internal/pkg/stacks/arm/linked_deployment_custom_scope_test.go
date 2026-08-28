package arm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const rgScopedCustomStackFixture = `{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {},
  "resources": [],
  "outputs": {}
}`

const subscriptionScopedCustomStackFixture = `{
  "$schema": "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {},
  "resources": [],
  "outputs": {}
}`

// At subscription scope ARM takes a nested deployment's target from its own
// resourceGroup/location, not the root's. With neither set, an RG-scoped child
// fails with InvalidScope and a subscription-scoped one has no location.
func TestGetCustomLinkedDeployments_SubscriptionScopeTargeting(t *testing.T) {
	rgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(rgScopedCustomStackFixture))
	}))
	defer rgServer.Close()

	subServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(subscriptionScopedCustomStackFixture))
	}))
	defer subServer.Close()

	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.DeploymentScope = app.StackDeploymentScopeSubscription
	inp.AppCfg.StackConfig.CustomNestedStacks = []config.CustomNestedStack{
		{Name: "rg_stack", TemplateURL: rgServer.URL + "/stack.json", Index: 0},
		{Name: "sub_stack", TemplateURL: subServer.URL + "/stack.json", Index: 1},
	}

	resources, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
	if err != nil {
		t.Fatalf("getCustomLinkedDeployments: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	rgDep := resources[0].(map[string]any)
	if got, present := rgDep["resourceGroup"]; !present || got != "[variables('installResourceGroupName')]" {
		t.Errorf("rg-scoped custom stack deployment: resourceGroup = %v (present=%v), want install RG", got, present)
	}
	if got, present := rgDep["location"]; present {
		t.Errorf("rg-scoped custom stack deployment must not set location, got %v", got)
	}

	subDep := resources[1].(map[string]any)
	if got, present := subDep["resourceGroup"]; present {
		t.Errorf("subscription-scoped custom stack deployment must not set resourceGroup, got %v", got)
	}
	if _, present := subDep["location"]; !present {
		t.Error("subscription-scoped custom stack deployment requires an explicit location")
	}
}
