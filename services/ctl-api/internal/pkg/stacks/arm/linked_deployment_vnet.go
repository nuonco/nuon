package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// defaultAzureVNetTemplateURL is the default VNet ARM template URL.
const defaultAzureVNetTemplateURL = "https://raw.githubusercontent.com/nuonco/sandboxes/main/azure-aks/vnet-template.json"

// vnetHoistAllowlist lists VNet template parameters that are intentionally
// customer-configurable. Only these are surfaced in the parent template.
var vnetHoistAllowlist = map[string]bool{
	"vnetCIDR":           true,
	"publicSubnet1CIDR":  true,
	"publicSubnet2CIDR":  true,
	"publicSubnet3CIDR":  true,
	"runnerSubnetCIDR":   true,
	"privateSubnet1CIDR": true,
	"privateSubnet2CIDR": true,
	"privateSubnet3CIDR": true,
}

func (t *Templates) getVNetLinkedDeployment(inp *stacks.TemplateInput) (map[string]any, map[string]ARMParameter, error) {
	templateURL := inp.VPCNestedStackTemplateURL
	if templateURL == "" {
		// No custom VNet template - build inline default VNet resources
		return t.getDefaultVNetDeployment(inp), nil, nil
	}

	// Custom VNet template — fetch and inspect declared parameters.
	armTmpl, err := fetchARMTemplate(templateURL)
	if err != nil {
		return nil, nil, fmt.Errorf("VNet linked deployment: %w", err)
	}

	// Nuon-managed parameters: always baked, never customer-facing.
	managedParams := map[string]any{
		"nuonInstallID": inp.Install.ID,
		"nuonOrgID":     inp.Runner.OrgID,
		"nuonAppID":     inp.Install.AppID,
		"location":      "[parameters('location')]",
		"commonTags":    "[variables('commonTags')]",
	}

	deploymentParams := map[string]any{}
	hoistedParams := map[string]ARMParameter{}

	for paramName, param := range armTmpl.Parameters {
		if val, ok := managedParams[paramName]; ok {
			// Nuon-managed — bake the value directly.
			deploymentParams[paramName] = map[string]any{"value": val}
		} else if vnetHoistAllowlist[paramName] {
			// Customer-configurable — hoist to parent template and reference.
			deploymentParams[paramName] = map[string]any{"value": fmt.Sprintf("[parameters('%s')]", paramName)}
			hp := ARMParameter{
				Type:         param.Type,
				DefaultValue: param.DefaultValue,
			}
			if param.Metadata != nil && param.Metadata.Description != "" {
				hp.Metadata = &ARMParameterMetadata{
					Description: param.Metadata.Description,
				}
			}
			hoistedParams[paramName] = hp
		}
		// Everything else is left to the template's own defaults.
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
		"commonTags":         map[string]any{"value": "[variables('commonTags')]"},
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
	// Core networking resources (Public IP, NAT Gateway, NSGs, Route Table, VNet).
	resources := []any{
		// Public IP for NAT Gateway
		map[string]any{
			"type":       "Microsoft.Network/publicIPAddresses",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-natgw-pip', parameters('nuonInstallID'))]",
			"location":   "[parameters('location')]",
			"tags":       "[parameters('commonTags')]",
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
			"tags":       "[parameters('commonTags')]",
			"sku":        map[string]any{"name": "Standard"},
			"properties": map[string]any{
				"publicIpAddresses": []any{
					map[string]any{
						"id": "[resourceId('Microsoft.Network/publicIPAddresses', format('{0}-natgw-pip', parameters('nuonInstallID')))]",
					},
				},
				"idleTimeoutInMinutes": 4,
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
			"tags":       "[parameters('commonTags')]",
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
			"tags":       "[parameters('commonTags')]",
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
			"tags":       "[parameters('commonTags')]",
			"properties": map[string]any{
				"disableBgpRoutePropagation": false,
			},
		},
		// VNet — subnets are declared as standalone child resources below so
		// that ARM does not delete externally-created subnets on re-deploy.
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet', parameters('nuonInstallID'))]",
			"location":   "[parameters('location')]",
			"tags":       "[parameters('commonTags')]",
			"properties": map[string]any{
				"addressSpace": map[string]any{
					"addressPrefixes": []string{"[parameters('vnetCIDR')]"},
				},
			},
			"dependsOn": []string{
				"[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]",
				"[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
				"[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
			},
		},
	}

	// Standalone subnet resources — each depends on the VNet and relevant
	// NSG/NAT resources. Because they are separate resources (not inline on
	// the VNet), ARM will not attempt to remove subnets that are not declared
	// in the template (e.g. subnets created by an AKS AGIC addon).
	resources = append(resources, t.buildDefaultSubnetResources()...)

	return map[string]any{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]any{
			"nuonInstallID": map[string]any{"type": "string"},
			"location":      map[string]any{"type": "string"},
			"commonTags":    map[string]any{"type": "object"},
			"vnetCIDR": map[string]any{"type": "string", "defaultValue": "10.128.0.0/16",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for this VNet."}},
			"publicSubnet1CIDR": map[string]any{"type": "string", "defaultValue": "10.128.0.0/26",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for the public subnet."}},
			"publicSubnet2CIDR": map[string]any{"type": "string", "defaultValue": "10.128.0.64/26",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for the public subnet in the second zone (optional)."}},
			"publicSubnet3CIDR": map[string]any{"type": "string", "defaultValue": "10.128.0.128/26",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for the public subnet in the third zone (optional)."}},
			"runnerSubnetCIDR": map[string]any{"type": "string", "defaultValue": "10.128.128.0/24",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for the dedicated private subnet for the runner."}},
			"privateSubnet1CIDR": map[string]any{"type": "string", "defaultValue": "10.128.130.0/24",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for the private subnet."}},
			"privateSubnet2CIDR": map[string]any{"type": "string", "defaultValue": "10.128.132.0/24",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for the private subnet in the second zone (optional)."}},
			"privateSubnet3CIDR": map[string]any{"type": "string", "defaultValue": "10.128.134.0/24",
				"metadata": map[string]any{"description": "IP range (CIDR notation) for the private subnet in the third zone (optional)."}},
		},
		"variables": map[string]any{
			"createPublicSubnet2":  "[not(empty(parameters('publicSubnet2CIDR')))]",
			"createPublicSubnet3":  "[not(empty(parameters('publicSubnet3CIDR')))]",
			"createPrivateSubnet2": "[not(empty(parameters('privateSubnet2CIDR')))]",
			"createPrivateSubnet3": "[not(empty(parameters('privateSubnet3CIDR')))]",
		},
		"resources": resources,
		"outputs":   buildDefaultVNetOutputs(),
	}
}

// buildDefaultSubnetResources returns subnets as standalone ARM child resources
// (Microsoft.Network/virtualNetworks/subnets) rather than inline VNet properties.
// This prevents ARM from deleting subnets that exist in Azure but are not declared
// in the template (e.g. an ingress-subnet created by an AKS AGIC addon).
func (t *Templates) buildDefaultSubnetResources() []any {
	serviceEndpoints := []map[string]any{
		{"service": "Microsoft.KeyVault"},
		{"service": "Microsoft.ContainerRegistry"},
	}

	vnetDep := []string{
		"[resourceId('Microsoft.Network/virtualNetworks', format('{0}-vnet', parameters('nuonInstallID')))]",
	}

	publicNSG := map[string]any{
		"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]",
	}
	privateNSG := map[string]any{
		"id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
	}
	natGW := map[string]any{
		"id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
	}

	return []any{
		// Public subnets
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks/subnets",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet/{0}-public-subnet-zone1', parameters('nuonInstallID'))]",
			"dependsOn":  vnetDep,
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('publicSubnet1CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup":              publicNSG,
				"natGateway":                        natGW,
			},
		},
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks/subnets",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet/{0}-public-subnet-zone2', parameters('nuonInstallID'))]",
			"dependsOn": []string{
				"[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone1', parameters('nuonInstallID')))]",
			},
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('publicSubnet2CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup":              publicNSG,
				"natGateway":                        natGW,
			},
		},
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks/subnets",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet/{0}-public-subnet-zone3', parameters('nuonInstallID'))]",
			"dependsOn": []string{
				"[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone2', parameters('nuonInstallID')))]",
			},
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('publicSubnet3CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup":              publicNSG,
				"natGateway":                        natGW,
			},
		},
		// Runner subnet
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks/subnets",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet/{0}-private-runner-subnet', parameters('nuonInstallID'))]",
			"dependsOn": []string{
				"[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone3', parameters('nuonInstallID')))]",
			},
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('runnerSubnetCIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup":              privateNSG,
				"natGateway":                        natGW,
				"serviceEndpoints":                  serviceEndpoints,
			},
		},
		// Private subnets
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks/subnets",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet/{0}-private-subnet-zone1', parameters('nuonInstallID'))]",
			"dependsOn": []string{
				"[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-runner-subnet', parameters('nuonInstallID')))]",
			},
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('privateSubnet1CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup":              privateNSG,
				"natGateway":                        natGW,
				"serviceEndpoints":                  serviceEndpoints,
			},
		},
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks/subnets",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet/{0}-private-subnet-zone2', parameters('nuonInstallID'))]",
			"dependsOn": []string{
				"[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID')))]",
			},
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('privateSubnet2CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup":              privateNSG,
				"natGateway":                        natGW,
				"serviceEndpoints":                  serviceEndpoints,
			},
		},
		map[string]any{
			"type":       "Microsoft.Network/virtualNetworks/subnets",
			"apiVersion": "2023-04-01",
			"name":       "[format('{0}-vnet/{0}-private-subnet-zone3', parameters('nuonInstallID'))]",
			"dependsOn": []string{
				"[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone2', parameters('nuonInstallID')))]",
			},
			"properties": map[string]any{
				"addressPrefix":                     "[parameters('privateSubnet3CIDR')]",
				"privateEndpointNetworkPolicies":    "Disabled",
				"privateLinkServiceNetworkPolicies": "Enabled",
				"networkSecurityGroup":              privateNSG,
				"natGateway":                        natGW,
				"serviceEndpoints":                  serviceEndpoints,
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
