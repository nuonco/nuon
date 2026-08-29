export default {
  title: 'Workflows/StepDetails/AwaitAzureDetails',
}

import { AwaitAzureDetails } from './AwaitAzureDetails'
import type { TAppInput, TAppSecretConfig } from '@/types'

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

const mockInputs = [
  {
    name: 'api_key',
    display_name: 'API key',
    description: 'Issued from your vendor account.',
    source: 'customer',
    required: true,
  },
  {
    name: 'log_level',
    description: 'Verbosity of the application logs.',
    source: 'customer',
    default: 'info',
  },
  {
    name: 'internal_toggle',
    source: 'vendor',
  },
] satisfies TAppInput[]

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      orgId="org-1"
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
      orgId="org-1"
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
      orgId="org-1"
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
      orgId="org-1"
      installId="install-1"
      azureLocation="eastus"
      secrets={mockSecrets}
    />
  </div>
)

export const WithCustomerInputs = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
      azureLocation="eastus"
      inputs={mockInputs}
    />
  </div>
)

// api_key already has a value, so the deploy command must not offer to replace it
// with a placeholder.
export const WithCustomerInputsAlreadySet = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
      azureLocation="eastus"
      inputs={mockInputs}
      setInputNames={new Set(['api_key'])}
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
      loading
    />
  </div>
)

export const TFModule = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
      azureLocation="eastus"
      azureSubscriptionId="00000000-0000-0000-0000-000000000000"
      tfProvider
    />
  </div>
)
