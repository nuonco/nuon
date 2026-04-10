package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// defaultAzureVNetTemplateURL is the default VNet ARM template URL.
const defaultAzureVNetTemplateURL = "https://raw.githubusercontent.com/nuonco/sandboxes/main/azure-aks/vnet-template.json"

func (t *Templates) getVNetLinkedDeployment(inp *stacks.TemplateInput) (map[string]any, map[string]ARMParameter, error) {
	templateURL := inp.VPCNestedStackTemplateURL
	if templateURL == "" {
		// No custom VNet template - build inline default VNet resources
		return t.getDefaultVNetDeployment(inp), nil, nil
	}

	// Custom VNet template - fetch, extract params, build linked deployment
	armTmpl, err := fetchARMTemplate(templateURL)
	if err != nil {
		return nil, nil, fmt.Errorf("VNet linked deployment: %w", err)
	}

	params, hoistedParams := extractARMParameters(armTmpl, ReservedParamNames)

	// Inject Nuon-managed parameters if template declares them
	nuonParams := map[string]string{
		"nuonInstallID": inp.Install.ID,
		"nuonOrgID":     inp.Runner.OrgID,
		"nuonAppID":     inp.Install.AppID,
		"location":      "[parameters('location')]",
	}
	deploymentParams := map[string]any{}
	for paramName := range params {
		if val, ok := nuonParams[paramName]; ok {
			deploymentParams[paramName] = map[string]any{"value": val}
		} else {
			// Hoisted parameter - reference from parent
			deploymentParams[paramName] = map[string]any{"value": fmt.Sprintf("[parameters('%s')]", paramName)}
		}
	}

	deployment := map[string]any{
		"type":       "Microsoft.Resources/deployments",
		"apiVersion": "2022-09-01",
		"name":       "vnetDeployment",
		"properties": map[string]any{
			"mode": "Incremental",
			"templateLink": map[string]any{
				"uri": templateURL,
			},
			"parameters": deploymentParams,
		},
	}

	return deployment, hoistedParams, nil
}

func (t *Templates) getDefaultVNetDeployment(inp *stacks.TemplateInput) map[string]any {
	installID := "[parameters('nuonInstallID')]"
	location := "[parameters('location')]"

	defaultParams := map[string]any{
		"vnetCIDR":           map[string]any{"value": "10.128.0.0/16"},
		"publicSubnet1CIDR":  map[string]any{"value": "10.128.0.0/26"},
		"publicSubnet2CIDR":  map[string]any{"value": "10.128.0.64/26"},
		"publicSubnet3CIDR":  map[string]any{"value": "10.128.0.128/26"},
		"runnerSubnetCIDR":   map[string]any{"value": "10.128.128.0/24"},
		"privateSubnet1CIDR": map[string]any{"value": "10.128.130.0/24"},
		"privateSubnet2CIDR": map[string]any{"value": "10.128.132.0/24"},
		"privateSubnet3CIDR": map[string]any{"value": "10.128.134.0/24"},
		"nuonInstallID":      map[string]any{"value": installID},
		"location":           map[string]any{"value": location},
	}

	deployment := map[string]any{
		"type":       "Microsoft.Resources/deployments",
		"apiVersion": "2022-09-01",
		"name":       "vnetDeployment",
		"properties": map[string]any{
			"mode": "Incremental",
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"parameters": defaultParams,
			"template":   t.getDefaultVNetTemplate(),
		},
	}

	return deployment
}

func (t *Templates) getDefaultVNetTemplate() map[string]any {
	return map[string]any{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]any{
			"nuonInstallID":      map[string]any{"type": "string"},
			"location":           map[string]any{"type": "string"},
			"vnetCIDR":           map[string]any{"type": "string", "defaultValue": "10.128.0.0/16"},
			"publicSubnet1CIDR":  map[string]any{"type": "string", "defaultValue": "10.128.0.0/26"},
			"publicSubnet2CIDR":  map[string]any{"type": "string", "defaultValue": "10.128.0.64/26"},
			"publicSubnet3CIDR":  map[string]any{"type": "string", "defaultValue": "10.128.0.128/26"},
			"runnerSubnetCIDR":   map[string]any{"type": "string", "defaultValue": "10.128.128.0/24"},
			"privateSubnet1CIDR": map[string]any{"type": "string", "defaultValue": "10.128.130.0/24"},
			"privateSubnet2CIDR": map[string]any{"type": "string", "defaultValue": "10.128.132.0/24"},
			"privateSubnet3CIDR": map[string]any{"type": "string", "defaultValue": "10.128.134.0/24"},
		},
		"variables": map[string]any{
			"createPublicSubnet2":  "[not(empty(parameters('publicSubnet2CIDR')))]",
			"createPublicSubnet3":  "[not(empty(parameters('publicSubnet3CIDR')))]",
			"createPrivateSubnet2": "[not(empty(parameters('privateSubnet2CIDR')))]",
			"createPrivateSubnet3": "[not(empty(parameters('privateSubnet3CIDR')))]",
		},
		"resources": []any{
			// Public IP for NAT Gateway
			map[string]any{
				"type":       "Microsoft.Network/publicIPAddresses",
				"apiVersion": "2023-04-01",
				"name":       "[format('{0}-natgw-pip', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"sku":        map[string]any{"name": "Standard"},
				"properties": map[string]any{
					"publicIPAllocationMethod": "Static",
				},
			},
			// NAT Gateway
			map[string]any{
				"type":       "Microsoft.Network/natGateways",
				"apiVersion": "2023-04-01",
				"name":       "[format('{0}-natgw', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"sku":        map[string]any{"name": "Standard"},
				"properties": map[string]any{
					"publicIpAddresses": []any{
						map[string]any{
							"id": "[resourceId('Microsoft.Network/publicIPAddresses', format('{0}-natgw-pip', parameters('nuonInstallID')))]",
						},
					},
				},
				"dependsOn": []string{
					"[resourceId('Microsoft.Network/publicIPAddresses', format('{0}-natgw-pip', parameters('nuonInstallID')))]",
				},
			},
			// Public NSG
			map[string]any{
				"type":       "Microsoft.Network/networkSecurityGroups",
				"apiVersion": "2023-04-01",
				"name":       "[format('{0}-public-nsg', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"properties": map[string]any{
					"securityRules": []any{
						map[string]any{
							"name": "Allow-All-Inbound",
							"properties": map[string]any{
								"description":              "Allow all inbound traffic from any source",
								"protocol":                 "*",
								"sourcePortRange":          "*",
								"destinationPortRange":     "*",
								"sourceAddressPrefix":      "*",
								"destinationAddressPrefix": "*",
								"access":                   "Allow",
								"priority":                 200,
								"direction":                "Inbound",
							},
						},
					},
				},
			},
			// Private NSG
			map[string]any{
				"type":       "Microsoft.Network/networkSecurityGroups",
				"apiVersion": "2023-04-01",
				"name":       "[format('{0}-private-nsg', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"properties": map[string]any{
					"securityRules": []any{},
				},
			},
			// Route Table
			map[string]any{
				"type":       "Microsoft.Network/routeTables",
				"apiVersion": "2023-04-01",
				"name":       "[format('{0}-private-routetable', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"properties": map[string]any{
					"disableBgpRoutePropagation": false,
				},
			},
			// VNet with subnets
			map[string]any{
				"type":       "Microsoft.Network/virtualNetworks",
				"apiVersion": "2023-04-01",
				"name":       "[format('{0}-vnet', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"properties": map[string]any{
					"addressSpace": map[string]any{
						"addressPrefixes": []string{"[parameters('vnetCIDR')]"},
					},
					"subnets": t.buildDefaultSubnets(),
				},
				"dependsOn": []string{
					"[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]",
					"[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
					"[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
			},
		},
		"outputs": buildDefaultVNetOutputs(),
	}
}

func (t *Templates) buildDefaultSubnets() []map[string]any {
	serviceEndpoints := []map[string]any{
		{"service": "Microsoft.KeyVault"},
		{"service": "Microsoft.ContainerRegistry"},
	}

	return []map[string]any{
		{
			"name": "[format('{0}-public-subnet-zone1', parameters('nuonInstallID'))]",
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('publicSubnet1CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup": map[string]any{
					"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]",
				},
				"natGateway": map[string]any{
					"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
			},
		},
		{
			"name": "[format('{0}-public-subnet-zone2', parameters('nuonInstallID'))]",
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('publicSubnet2CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup": map[string]any{
					"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]",
				},
				"natGateway": map[string]any{
					"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
			},
		},
		{
			"name": "[format('{0}-public-subnet-zone3', parameters('nuonInstallID'))]",
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('publicSubnet3CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup": map[string]any{
					"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]",
				},
				"natGateway": map[string]any{
					"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
			},
		},
		{
			"name": "[format('{0}-private-runner-subnet', parameters('nuonInstallID'))]",
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('runnerSubnetCIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup": map[string]any{
					"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
				},
				"natGateway": map[string]any{
					"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
				"serviceEndpoints": serviceEndpoints,
			},
		},
		{
			"name": "[format('{0}-private-subnet-zone1', parameters('nuonInstallID'))]",
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('privateSubnet1CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup": map[string]any{
					"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
				},
				"natGateway": map[string]any{
					"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
				"serviceEndpoints": serviceEndpoints,
			},
		},
		{
			"name": "[format('{0}-private-subnet-zone2', parameters('nuonInstallID'))]",
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('privateSubnet2CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup": map[string]any{
					"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
				},
				"natGateway": map[string]any{
					"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
				"serviceEndpoints": serviceEndpoints,
			},
		},
		{
			"name": "[format('{0}-private-subnet-zone3', parameters('nuonInstallID'))]",
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('privateSubnet3CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup": map[string]any{
					"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
				},
				"natGateway": map[string]any{
					"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
				},
				"serviceEndpoints": serviceEndpoints,
			},
		},
	}
}

func buildDefaultVNetOutputs() map[string]any {
	return map[string]any{
		"vnetId":             map[string]any{"type": "string", "value": "[resourceId('Microsoft.Network/virtualNetworks', format('{0}-vnet', parameters('nuonInstallID')))]"},
		"vnetName":           map[string]any{"type": "string", "value": "[format('{0}-vnet', parameters('nuonInstallID'))]"},
		"runnerSubnetId":     map[string]any{"type": "string", "value": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-runner-subnet', parameters('nuonInstallID')))]"},
		"runnerSubnetName":   map[string]any{"type": "string", "value": "[format('{0}-private-runner-subnet', parameters('nuonInstallID'))]"},
		"publicSubnet1Id":    map[string]any{"type": "string", "value": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone1', parameters('nuonInstallID')))]"},
		"publicSubnet1Name":  map[string]any{"type": "string", "value": "[format('{0}-public-subnet-zone1', parameters('nuonInstallID'))]"},
		"publicSubnet2Id":    map[string]any{"type": "string", "value": "[if(variables('createPublicSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone2', parameters('nuonInstallID'))), '')]"},
		"publicSubnet2Name":  map[string]any{"type": "string", "value": "[if(variables('createPublicSubnet2'), format('{0}-public-subnet-zone2', parameters('nuonInstallID')), '')]"},
		"publicSubnet3Id":    map[string]any{"type": "string", "value": "[if(variables('createPublicSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone3', parameters('nuonInstallID'))), '')]"},
		"publicSubnet3Name":  map[string]any{"type": "string", "value": "[if(variables('createPublicSubnet3'), format('{0}-public-subnet-zone3', parameters('nuonInstallID')), '')]"},
		"privateSubnet1Id":   map[string]any{"type": "string", "value": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID')))]"},
		"privateSubnet1Name": map[string]any{"type": "string", "value": "[format('{0}-private-subnet-zone1', parameters('nuonInstallID'))]"},
		"privateSubnet2Id":   map[string]any{"type": "string", "value": "[if(variables('createPrivateSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone2', parameters('nuonInstallID'))), '')]"},
		"privateSubnet2Name": map[string]any{"type": "string", "value": "[if(variables('createPrivateSubnet2'), format('{0}-private-subnet-zone2', parameters('nuonInstallID')), '')]"},
		"privateSubnet3Id":   map[string]any{"type": "string", "value": "[if(variables('createPrivateSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone3', parameters('nuonInstallID'))), '')]"},
		"privateSubnet3Name": map[string]any{"type": "string", "value": "[if(variables('createPrivateSubnet3'), format('{0}-private-subnet-zone3', parameters('nuonInstallID')), '')]"},
		"publicSubnetIds":    map[string]any{"type": "string", "value": "[join(filter(createArray(resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone1', parameters('nuonInstallID'))), if(variables('createPublicSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone2', parameters('nuonInstallID'))), ''), if(variables('createPublicSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone3', parameters('nuonInstallID'))), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"},
		"publicSubnetNames":  map[string]any{"type": "string", "value": "[join(filter(createArray(format('{0}-public-subnet-zone1', parameters('nuonInstallID')), if(variables('createPublicSubnet2'), format('{0}-public-subnet-zone2', parameters('nuonInstallID')), ''), if(variables('createPublicSubnet3'), format('{0}-public-subnet-zone3', parameters('nuonInstallID')), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"},
		"privateSubnetIds":   map[string]any{"type": "string", "value": "[join(filter(createArray(resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID'))), if(variables('createPrivateSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone2', parameters('nuonInstallID'))), ''), if(variables('createPrivateSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone3', parameters('nuonInstallID'))), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"},
		"privateSubnetNames": map[string]any{"type": "string", "value": "[join(filter(createArray(format('{0}-private-subnet-zone1', parameters('nuonInstallID')), if(variables('createPrivateSubnet2'), format('{0}-private-subnet-zone2', parameters('nuonInstallID')), ''), if(variables('createPrivateSubnet3'), format('{0}-private-subnet-zone3', parameters('nuonInstallID')), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"},
	}
}
