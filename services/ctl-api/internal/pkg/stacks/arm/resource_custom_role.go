package arm

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) getCustomRoleDeployment(inp *stacks.TemplateInput) map[string]any {
	_ = inp
	return map[string]any{
		"type":           "Microsoft.Resources/deployments",
		"apiVersion":     "2022-09-01",
		"name":           "[format('{0}-custom-role-deployment', parameters('nuonInstallID'))]",
		"subscriptionId": "[subscription().subscriptionId]",
		"location":       "[resourceGroup().location]",
		"dependsOn":      []string{"runnerDeployment"},
		"properties": map[string]any{
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"mode": "Incremental",
			"parameters": map[string]any{
				"nuonInstallID": map[string]any{"value": "[parameters('nuonInstallID')]"},
				"principalID":   map[string]any{"value": "[reference('runnerDeployment').outputs.vmssPrincipalId.value]"},
			},
			"template": map[string]any{
				"$schema":        "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
				"parameters": map[string]any{
					"nuonInstallID": map[string]any{"type": "string"},
					"principalID":   map[string]any{"type": "string"},
				},
				"resources": []map[string]any{
					{
						"type":       "Microsoft.Authorization/roleDefinitions",
						"apiVersion": "2022-04-01",
						"name":       "[guid(subscription().id, format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID')))]",
						"properties": map[string]any{
							"roleName":    "[format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID'))]",
							"description": "Custom role to allow assuming other trusted roles",
							"assignableScopes": []string{
								"[subscription().id]",
							},
							"permissions": []map[string]any{
								{
									"actions":        []string{"*/register/action"},
									"notActions":     []string{},
									"dataActions":    []string{},
									"notDataActions": []string{},
								},
							},
						},
					},
					{
						"type":       "Microsoft.Authorization/roleAssignments",
						"apiVersion": "2022-04-01",
						"name":       "[guid(subscription().id, parameters('principalID'), 'CustomRunnerRole')]",
						"properties": map[string]any{
							"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID'))))]",
							"principalId":      "[parameters('principalID')]",
							"principalType":    "ServicePrincipal",
						},
						"dependsOn": []string{
							"[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID'))))]",
						},
					},
				},
			},
		},
	}
}
