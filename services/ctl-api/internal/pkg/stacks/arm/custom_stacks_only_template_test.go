package arm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const customStacksOnlyTestTemplate = `{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "vnetId": { "type": "string" },
    "setting": { "type": "string", "defaultValue": "default" }
  },
  "resources": [],
  "outputs": {
    "resourceName": { "type": "string", "value": "[parameters('setting')]" }
  }
}`

func TestAzureCustomStacksOnlyTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(customStacksOnlyTestTemplate))
	}))
	defer server.Close()

	inp := armCustomStackInput(t, server.URL, map[string]string{"setting": "sensitive-value"})
	inp.CustomStacksOnly = true
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"route53_zones": {"setting": "{{.nuon.install.inputs.setting}}"},
	}
	inp.AppCfg.InputConfig.AppInputs = []app.AppInput{
		{Name: "setting", Source: app.AppInputSourceCustomer},
	}

	templates := &Templates{cfg: &internal.Config{}}
	tmpl, outputMap, inputParametersMap, err := templates.getAzureCustomStacksOnlyTemplate(inp)
	require.NoError(t, err)

	assert.Equal(t, subscriptionTemplateSchema, tmpl.Schema)
	assert.Contains(t, tmpl.Parameters, customStacksResourceGroupParameter)
	assert.Contains(t, tmpl.Parameters, "location")
	assert.Contains(t, tmpl.Parameters, "vnetId")
	assert.NotContains(t, tmpl.Parameters, "setting")
	assert.Contains(t, tmpl.Parameters, "Route53ZonesSetting")
	assert.Equal(t, "[parameters('"+customStacksResourceGroupParameter+"')]", tmpl.Variables[installRGVarName])

	require.Len(t, tmpl.Resources, 1)
	deployment := tmpl.Resources[0].(map[string]any)
	assert.Equal(t, "[variables('"+installRGVarName+"')]", deployment["resourceGroup"])
	assert.Empty(t, deployment["dependsOn"])
	assert.Equal(t, "[parameters('vnetId')]", armDeploymentParamValue(t, deployment, "vnetId"))
	assert.Equal(t, "[parameters('Route53ZonesSetting')]", armDeploymentParamValue(t, deployment, "setting"))
	assert.Equal(t, "setting", inputParametersMap["route53_zones"]["Route53ZonesSetting"])

	flatName := outputMap["route53_zones"]["resourceName"]
	require.NotEmpty(t, flatName)
	assert.Equal(t, "[string(reference('"+deployment["name"].(string)+"').outputs.resourceName.value)]", tmpl.Outputs[flatName].Value)
}

func TestAzureCustomStacksOnlyTemplateBakesVendorInputs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(customStacksOnlyTestTemplate))
	}))
	defer server.Close()

	inp := armCustomStackInput(t, server.URL, map[string]string{"setting": "vendor-value"})
	inp.CustomStacksOnly = true
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"route53_zones": {"setting": "{{.nuon.install.inputs.setting}}"},
	}
	inp.AppCfg.InputConfig.AppInputs = []app.AppInput{
		{Name: "setting", Source: app.AppInputSourceVendor},
	}

	templates := &Templates{cfg: &internal.Config{}}
	tmpl, _, inputParametersMap, err := templates.getAzureCustomStacksOnlyTemplate(inp)
	require.NoError(t, err)

	assert.NotContains(t, tmpl.Parameters, "Route53ZonesSetting")
	deployment := tmpl.Resources[0].(map[string]any)
	assert.Equal(t, "vendor-value", armDeploymentParamValue(t, deployment, "setting"))
	assert.Empty(t, inputParametersMap)
}

func TestAzureCustomStacksOnlyTemplateTargetsSubscriptionScopedChildren(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
		  "$schema": "` + subscriptionTemplateSchema + `",
		  "contentVersion": "1.0.0.0",
		  "resources": [],
		  "outputs": {}
		}`))
	}))
	defer server.Close()

	inp := armCustomStackInput(t, server.URL, nil)
	inp.CustomStacksOnly = true

	templates := &Templates{cfg: &internal.Config{}}
	tmpl, _, _, err := templates.getAzureCustomStacksOnlyTemplate(inp)
	require.NoError(t, err)

	require.Len(t, tmpl.Resources, 1)
	deployment := tmpl.Resources[0].(map[string]any)
	assert.NotContains(t, deployment, "resourceGroup")
	assert.Equal(t, "[variables('location')]", deployment["location"])
}
