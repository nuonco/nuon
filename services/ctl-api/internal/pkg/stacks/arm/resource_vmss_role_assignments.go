package arm

// runnerGrantContext is how a runner grant addresses the VMSS system identity and
// what it waits on.
//
// At resource-group scope the grants sit in the root beside the runner deployment
// and read its output directly. Inside runnerGrantsDeployment they cannot: inner
// expression evaluation hides the root, so the principal arrives as a parameter
// and the wrapper holds the dependency on the runner.
//
// Everything else in these assignments — the guid() names and the Key Vault
// scope — is left alone on purpose. Inside the wrapper resourceGroup() resolves to
// the install group again, so those expressions are already correct, and rewriting
// a name would change the assignment GUID and fail redeploys with
// RoleAssignmentExists.
type runnerGrantContext struct {
	principalID string
	dependsOn   []string
}

func runnerGrantContextFor(scope armScope) runnerGrantContext {
	if scope.subscription {
		return runnerGrantContext{principalID: "[parameters('principalId')]"}
	}
	return runnerGrantContext{
		principalID: "[reference('runnerDeployment').outputs.vmssPrincipalId.value]",
		dependsOn:   []string{"runnerDeployment"},
	}
}

// apply sets the principal and, only when there is one, the dependency. An empty
// dependsOn key would be noise inside the wrapper.
func (c runnerGrantContext) apply(assignment map[string]any) map[string]any {
	assignment["properties"].(map[string]any)["principalId"] = c.principalID
	if len(c.dependsOn) > 0 {
		assignment["dependsOn"] = c.dependsOn
	}
	return assignment
}

func (t *Templates) getKeyVaultRoleAssignment(ctx runnerGrantContext) map[string]any {
	kvName := keyVaultNameInner

	return ctx.apply(map[string]any{
		"type":       "Microsoft.Authorization/roleAssignments",
		"apiVersion": "2022-04-01",
		"name":       "[guid(resourceId('Microsoft.KeyVault/vaults', " + kvName + "), resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'KeyVaultSecretsUser')]",
		"scope":      "[resourceId('Microsoft.KeyVault/vaults', " + kvName + ")]",
		"properties": map[string]any{
			"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '4633458b-17de-408a-b874-0445c86b69e6')]",
			"principalType":    "ServicePrincipal",
		},
	})
}

func (t *Templates) getVMSSRoleAssignments(ctx runnerGrantContext) []any {
	return []any{
		ctx.apply(map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'Contributor')]",
			"properties": map[string]any{
				"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b24988ac-6180-42a0-ab88-20f7382dd24c')]",
				"principalType":    "ServicePrincipal",
			},
		}),
		ctx.apply(map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'RoleBasedAccessControlAdministrator')]",
			"properties": map[string]any{
				"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'f58310d9-a9f6-439a-9e8d-f62e7b41a168')]",
				"principalType":    "ServicePrincipal",
			},
		}),
		ctx.apply(map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'AzureKubernetesServiceRBACClusterAdmin')]",
			"properties": map[string]any{
				"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b')]",
				"principalType":    "ServicePrincipal",
			},
		}),
	}
}

// getACRRoleAssignments grants the runner's system identity pull/push on the
// install's registry at resource-group scope. Image sync runs as the ambient
// identity (see pkg/azure/acr), so this stays on the system identity even when
// per-operation identities hold all other deploy grants.
func (t *Templates) getACRRoleAssignments(ctx runnerGrantContext) []any {
	vmssRef := "resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID')))"

	roles := []struct{ name, guid string }{
		{"AcrPull", "7f951dda-4ed3-4680-a7ca-43fe172d538d"},
		{"AcrPush", "8311e382-0749-4cb8-b61a-304f252e45ec"},
	}

	var assignments []any
	for _, r := range roles {
		assignments = append(assignments, ctx.apply(map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       "[guid(resourceGroup().id, " + vmssRef + ", '" + r.name + "')]",
			"properties": map[string]any{
				"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '" + r.guid + "')]",
				"principalType":    "ServicePrincipal",
			},
		}))
	}
	return assignments
}
