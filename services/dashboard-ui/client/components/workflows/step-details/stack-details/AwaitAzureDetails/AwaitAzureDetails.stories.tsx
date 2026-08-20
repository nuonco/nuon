export default {
  title: 'Workflows/StepDetails/AwaitAzureDetails',
}

import { AwaitAzureDetails } from './AwaitAzureDetails'
import type { TAppSecretConfig } from '@/types'

const mockStack = {
  versions: [
    {
      template_url: 'https://storage.azure.com/template.json',
    },
  ],
} as any

const mockStackWithQuickLink = {
  versions: [
    {
      template_url: 'https://storage.azure.com/template.json',
      quick_link_url:
        'https://portal.azure.com/#create/Microsoft.Template/uri/https%3A%2F%2Fstorage.azure.com%2Ftemplate-quicklink.json/createUIDefinitionUri/https%3A%2F%2Fstorage.azure.com%2Ftemplate-uidef.json',
    },
  ],
} as any

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

const mockSecrets = [
  {
    name: 'database_password',
    display_name: 'Database password',
    description: 'Password used by the application database.',
    required: true,
    auto_generate: false,
  },
] satisfies TAppSecretConfig[]

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      installId="install-1"
      azureLocation="eastus"
    />
  </div>
)

export const WithQuickLink = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStackWithQuickLink}
      step={mockStep}
      installId="install-1"
      azureLocation="eastus"
      deploymentScope="subscription"
    />
  </div>
)

export const QuickLinkHiddenAtResourceGroupScope = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStackWithQuickLink}
      step={mockStep}
      installId="install-1"
      azureLocation="eastus"
      deploymentScope="resource_group"
    />
  </div>
)

export const WithApplicationSecrets = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      installId="install-1"
      azureLocation="eastus"
      secrets={mockSecrets}
    />
  </div>
)

// The stack declares the resource group, the Key Vault and the secrets itself,
// so none of those prerequisites render and the deploy commands are sub-scoped.
export const SubscriptionScopeHidesPrerequisites = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStackWithQuickLink}
      step={mockStep}
      installId="install-1"
      azureLocation="eastus"
      secrets={mockSecrets}
      deploymentScope="subscription"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      installId="install-1"
      loading
    />
  </div>
)
