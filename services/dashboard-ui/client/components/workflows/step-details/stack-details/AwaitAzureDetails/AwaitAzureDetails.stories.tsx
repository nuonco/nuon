export default {
  title: 'Workflows/StepDetails/AwaitAzureDetails',
}

import { AwaitAzureDetails, AwaitAzureDetailsSkeleton } from './AwaitAzureDetails'

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

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetailsSkeleton />
  </div>
)
