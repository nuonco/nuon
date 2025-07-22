targetScope = 'subscription'

param nuonInstallID string
param principalID string

resource resourceProviderRegisterRoleDefinition 'Microsoft.Authorization/roleDefinitions@2022-04-01' = {
  name: guid(subscription().id, '${nuonInstallID}-runner-resource-provider-register-role')
  properties: {
    roleName: '${nuonInstallID}-runner-resource-provider-register-role'
    description: 'Custom role to allow assuming other trusted roles'
    assignableScopes: [
      subscription().id
    ]
    permissions: [
      {
        actions: [
          '*/register/action'
        ]
        notActions: []
        dataActions: []
        notDataActions: []
      }
    ]
  }
}

resource resourceProviderRegisterRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(subscription().id, principalID, 'CustomRunnerRole')
  properties: {
    roleDefinitionId: resourceProviderRegisterRoleDefinition.id
    principalId: principalID
    principalType: 'ServicePrincipal'
  }
}
