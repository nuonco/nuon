package arm

func (t *Templates) getKeyVaultRoleAssignment() map[string]any {
	principalId := "[reference('runnerDeployment').outputs.vmssPrincipalId.value]"
	kvName := "take(format('{0}', parameters('nuonInstallID')), 24)"

	return map[string]any{
		"type":       "Microsoft.Authorization/roleAssignments",
		"apiVersion": "2022-04-01",
		"name":       "[guid(resourceId('Microsoft.KeyVault/vaults', " + kvName + "), resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'KeyVaultSecretsUser')]",
		"scope":      "[resourceId('Microsoft.KeyVault/vaults', " + kvName + ")]",
		"dependsOn": []string{
			"runnerDeployment",
		},
		"properties": map[string]any{
			"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '4633458b-17de-408a-b874-0445c86b69e6')]",
			"principalId":      principalId,
			"principalType":    "ServicePrincipal",
		},
	}
}

func (t *Templates) getVMSSRoleAssignments() []any {
	principalId := "[reference('runnerDeployment').outputs.vmssPrincipalId.value]"

	return []any{
		map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'Contributor')]",
			"dependsOn":  []string{"runnerDeployment"},
			"properties": map[string]any{
				"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b24988ac-6180-42a0-ab88-20f7382dd24c')]",
				"principalId":      principalId,
				"principalType":    "ServicePrincipal",
			},
		},
		map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'RoleBasedAccessControlAdministrator')]",
			"dependsOn":  []string{"runnerDeployment"},
			"properties": map[string]any{
				"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'f58310d9-a9f6-439a-9e8d-f62e7b41a168')]",
				"principalId":      principalId,
				"principalType":    "ServicePrincipal",
			},
		},
		map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'AzureKubernetesServiceRBACClusterAdmin')]",
			"dependsOn":  []string{"runnerDeployment"},
			"properties": map[string]any{
				"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b')]",
				"principalId":      principalId,
				"principalType":    "ServicePrincipal",
			},
		},
	}
}
