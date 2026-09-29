package arm

import "fmt"

func (t *Templates) getCustomDeploymentRoleAssignment(id customDeploymentIdentity, installID string, scope armScope) map[string]any {
	roleKey := customStackRoleKey(installID, id.SanitizedName)
	deploymentName := customStackRoleDeploymentName(installID, id.SanitizedName)

	return map[string]any{
		"type":           "Microsoft.Resources/deployments",
		"apiVersion":     "2022-09-01",
		"name":           deploymentName,
		"subscriptionId": "[subscription().subscriptionId]",
		"location":       scope.locationExpr(),
		"dependsOn":      []string{id.DeploymentName},
		"properties": map[string]any{
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"mode": "Incremental",
			"parameters": map[string]any{
				"roleKey":     map[string]any{"value": roleKey},
				"principalID": map[string]any{"value": fmt.Sprintf("[reference('%s').outputs.%s.value]", id.DeploymentName, id.PrincipalIDOutput)},
			},
			"template": map[string]any{
				"$schema":        "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
				"parameters": map[string]any{
					"roleKey":     map[string]any{"type": "string"},
					"principalID": map[string]any{"type": "string"},
				},
				"resources": []map[string]any{
					{
						"type":       "Microsoft.Authorization/roleDefinitions",
						"apiVersion": "2022-04-01",
						"name":       "[guid(subscription().id, format('{0}-register-role', parameters('roleKey')))]",
						"properties": map[string]any{
							"roleName":    "[format('{0}-register-role', parameters('roleKey'))]",
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
						"name":       "[guid(subscription().id, parameters('principalID'), parameters('roleKey'))]",
						"dependsOn": []string{
							"[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-register-role', parameters('roleKey'))))]",
						},
						"properties": map[string]any{
							"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-register-role', parameters('roleKey'))))]",
							"principalId":      "[parameters('principalID')]",
							"principalType":    "ServicePrincipal",
						},
					},
				},
			},
		},
	}
}
