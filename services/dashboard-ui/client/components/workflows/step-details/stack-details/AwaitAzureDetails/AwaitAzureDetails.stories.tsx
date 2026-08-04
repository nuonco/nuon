export default {
  title: 'Workflows/StepDetails/AwaitAzureDetails',
}

import {
  AwaitAzureDetails,
  AwaitAzureDetailsSkeleton,
} from './AwaitAzureDetails'
import type { TAppSecretConfig } from '@/types'

const mockStack = {
  versions: [
    {
      template_url: 'https://storage.azure.com/template.json',
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

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetailsSkeleton />
  </div>
)
