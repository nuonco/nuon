package arm

import (
	"encoding/json"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/stretchr/testify/require"
)

func TestAzureTelemetryContract(t *testing.T) {
	for _, subscription := range []bool{false, true} {
		for _, tc := range []struct {
			name      string
			custom    string
			local     bool
			supported bool
		}{
			{"built-in", "", false, true},
			{"custom", `{"parameters":{"userAssignedIdentities":{"type":"object"},"enableTelemetryIngress":{"type":"bool","defaultValue":true}},"outputs":{"telemetryEndpoint":{"type":"string"}},"resources":[]}`, false, true},
			{"older custom", `{"parameters":{"userAssignedIdentities":{"type":"object"}},"resources":[]}`, false, false},
			{"missing output", `{"parameters":{"userAssignedIdentities":{"type":"object"},"enableTelemetryIngress":{"type":"bool"}},"resources":[]}`, false, false},
			{"missing toggle", `{"parameters":{"userAssignedIdentities":{"type":"object"}},"outputs":{"telemetryEndpoint":{"type":"string"}},"resources":[]}`, false, false},
			{"wrong toggle type", `{"parameters":{"userAssignedIdentities":{"type":"object"},"enableTelemetryIngress":{"type":"string"}},"outputs":{"telemetryEndpoint":{"type":"string"}},"resources":[]}`, false, false},
			{"local runner", "", true, false},
		} {
			t.Run(tc.name+map[bool]string{false: "/resource-group", true: "/subscription"}[subscription], func(t *testing.T) {
				inp := minimalTemplateInput()
				if subscription {
					inp = subscriptionTemplateInput()
				}
				if tc.custom != "" {
					inp.RunnerNestedStackTemplateURL = runnerTemplateServer(t, tc.custom)
				}
				tmpl, err := (&Templates{cfg: &internal.Config{UseLocalRunners: tc.local}}).getAzureTemplate(inp)
				require.NoError(t, err)
				raw, err := json.Marshal(tmpl)
				require.NoError(t, err)
				if !tc.supported {
					require.NotContains(t, tmpl.Parameters, "enableTelemetryIngress")
					require.NotContains(t, tmpl.Outputs, "telemetryEndpoint")
					require.NotContains(t, string(raw), "TELEMETRY_ENDPOINT")
					return
				}
				require.Equal(t, "bool", tmpl.Parameters["enableTelemetryIngress"].Type)
				require.Equal(t, true, tmpl.Parameters["enableTelemetryIngress"].DefaultValue)
				require.Equal(t, "[reference('runnerDeployment').outputs.telemetryEndpoint.value]", tmpl.Outputs["telemetryEndpoint"].Value)
				require.Contains(t, string(raw), `\"telemetry_endpoint\": \"$TELEMETRY_ENDPOINT\"`)
				for _, resource := range tmpl.Resources {
					r := resource.(map[string]any)
					if r["name"] == "runnerDeployment" {
						params := r["properties"].(map[string]any)["parameters"].(map[string]any)
						require.Equal(t, map[string]any{"value": "[parameters('enableTelemetryIngress')]"}, params["enableTelemetryIngress"])
					}
					if r["type"] == "Microsoft.Resources/deploymentScripts" || r["name"] == phoneHomeDeploymentName {
						require.Contains(t, r["dependsOn"], "runnerDeployment")
					}
				}
			})
		}
	}
}

func TestDefaultRunnerTelemetryResources(t *testing.T) {
	tmpl := (&Templates{}).getDefaultRunnerTemplate(nil, "Standard_D4s_v3")
	resources := tmpl["resources"].([]any)
	require.Len(t, resources, 2)
	vmss := resources[0].(map[string]any)
	lb := resources[1].(map[string]any)
	require.Equal(t, "[parameters('enableTelemetryIngress')]", lb["condition"])
	require.Equal(t, "Microsoft.Network/loadBalancers", lb["type"])
	require.Equal(t, map[string]any{"name": "Standard", "tier": "Regional"}, lb["sku"])
	props := lb["properties"].(map[string]any)
	frontend := props["frontendIPConfigurations"].([]any)[0].(map[string]any)["properties"].(map[string]any)
	require.NotContains(t, frontend, "publicIPAddress")
	require.Equal(t, map[string]any{"id": "[parameters('runnerSubnetId')]"}, frontend["subnet"])
	rule := props["loadBalancingRules"].([]any)[0].(map[string]any)["properties"].(map[string]any)
	require.Equal(t, 4318, rule["frontendPort"])
	require.Equal(t, 4318, rule["backendPort"])
	require.Equal(t, "Tcp", rule["protocol"])
	probe := props["probes"].([]any)[0].(map[string]any)["properties"].(map[string]any)
	require.Equal(t, 4318, probe["port"])
	profile := vmss["properties"].(map[string]any)["virtualMachineProfile"].(map[string]any)
	health := profile["extensionProfile"].(map[string]any)["extensions"].([]map[string]any)[0]["properties"].(map[string]any)["settings"].(map[string]any)
	require.Equal(t, 9999, health["port"])
	nic := profile["networkProfile"].(map[string]any)["networkInterfaceConfigurations"].([]map[string]any)[0]
	ip := nic["properties"].(map[string]any)["ipConfigurations"].([]map[string]any)[0]["properties"].(map[string]any)
	require.Equal(t, "[if(parameters('enableTelemetryIngress'), createArray(createObject('id', resourceId('Microsoft.Network/loadBalancers/backendAddressPools', format('{0}-telemetry', parameters('nuonInstallID')), 'runner'))), createArray())]", ip["loadBalancerBackendAddressPools"])
	endpoint := tmpl["outputs"].(map[string]any)["telemetryEndpoint"].(map[string]any)["value"]
	require.Equal(t, "[if(parameters('enableTelemetryIngress'), format('http://{0}:4318', first(reference(resourceId('Microsoft.Network/loadBalancers', format('{0}-telemetry', parameters('nuonInstallID'))), '2023-09-01').frontendIPConfigurations).properties.privateIPAddress), '')]", endpoint)
}
