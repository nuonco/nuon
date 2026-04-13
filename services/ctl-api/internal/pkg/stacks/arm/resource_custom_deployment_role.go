package arm

import "fmt"

// getCustomDeploymentRoleAssignment creates a subscription-level nested
// deployment that defines a custom role with */register/action and assigns it
// to the managed identity created inside a custom linked deployment.
//
// The identity's principalId is read from the linked deployment's
// "identityPrincipalId" output, so the custom nested stack template must
// expose that output.
func (t *Templates) getCustomDeploymentRoleAssignment(id customDeploymentIdentity) map[string]any {
	deploymentName := fmt.Sprintf("%s-identity-role", id.DeploymentName)

	return map[string]any{
		"type":           "Microsoft.Resources/deployments",
		"apiVersion":     "2022-09-01",
		"name":           deploymentName,
		"subscriptionId": "[subscription().subscriptionId]",
		"location":       "[resourceGroup().location]",
		"dependsOn":      []string{id.DeploymentName},
		"properties": map[string]any{
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"mode": "Incremental",
			"parameters": map[string]any{
				"deploymentName": map[string]any{"value": id.DeploymentName},
				"principalID":    map[string]any{"value": fmt.Sprintf("[reference('%s').outputs.identityPrincipalId.value]", id.DeploymentName)},
			},
			"template": map[string]any{
				"$schema":        "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
				"parameters": map[string]any{
					"deploymentName": map[string]any{"type": "string"},
					"principalID":    map[string]any{"type": "string"},
				},
				"resources": []map[string]any{
					{
						"type":       "Microsoft.Authorization/roleDefinitions",
						"apiVersion": "2022-04-01",
						"name":       "[guid(subscription().id, format('{0}-register-role', parameters('deploymentName')))]",
						"properties": map[string]any{
							"roleName":    "[format('{0}-register-role', parameters('deploymentName'))]",
							"description": "Custom role to register Azure resource providers",
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
						"name":       "[guid(subscription().id, parameters('principalID'), parameters('deploymentName'))]",
						"dependsOn": []string{
							"[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-register-role', parameters('deploymentName'))))]",
						},
						"properties": map[string]any{
							"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-register-role', parameters('deploymentName'))))]",
							"principalId":      "[parameters('principalID')]",
							"principalType":    "ServicePrincipal",
						},
					},
				},
			},
		},
	}
}
