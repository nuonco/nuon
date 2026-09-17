package arm

func telemetryIngressParameters() map[string]ARMParameter {
	return map[string]ARMParameter{
		"enableTelemetryIngress": {
			Type: "bool", DefaultValue: true,
			Metadata: &ARMParameterMetadata{Description: "Provision a private OTLP HTTP endpoint for this install (additional Azure charges apply)."},
		},
	}
}

func getRunnerTelemetryLoadBalancer() map[string]any {
	return map[string]any{
		"condition": "[parameters('enableTelemetryIngress')]",
		"type":      "Microsoft.Network/loadBalancers", "apiVersion": "2023-09-01",
		"name":     "[format('{0}-telemetry', parameters('nuonInstallID'))]",
		"location": "[parameters('location')]", "tags": "[parameters('commonTags')]",
		"sku": map[string]any{"name": "Standard", "tier": "Regional"},
		"properties": map[string]any{
			"frontendIPConfigurations": []any{map[string]any{
				"name": "telemetry",
				"properties": map[string]any{
					"privateIPAllocationMethod": "Dynamic",
					"subnet":                    map[string]any{"id": "[parameters('runnerSubnetId')]"},
				},
			}},
			"backendAddressPools": []any{map[string]any{"name": "runner"}},
			"probes": []any{map[string]any{
				"name":       "otlp",
				"properties": map[string]any{"protocol": "Tcp", "port": 4318, "intervalInSeconds": 5, "numberOfProbes": 2},
			}},
			"loadBalancingRules": []any{map[string]any{
				"name": "otlp",
				"properties": map[string]any{
					"frontendIPConfiguration": map[string]any{"id": "[resourceId('Microsoft.Network/loadBalancers/frontendIPConfigurations', format('{0}-telemetry', parameters('nuonInstallID')), 'telemetry')]"},
					"backendAddressPool":      map[string]any{"id": "[resourceId('Microsoft.Network/loadBalancers/backendAddressPools', format('{0}-telemetry', parameters('nuonInstallID')), 'runner')]"},
					"probe":                   map[string]any{"id": "[resourceId('Microsoft.Network/loadBalancers/probes', format('{0}-telemetry', parameters('nuonInstallID')), 'otlp')]"},
					"protocol":                "Tcp", "frontendPort": 4318, "backendPort": 4318,
					"disableOutboundSnat": true, "enableFloatingIP": false,
				},
			}},
		},
	}
}
