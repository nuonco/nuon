export default {
  title: 'Workflows/StepDetails/AwaitGCPDetails',
}

import { AwaitGCPDetails, AwaitGCPDetailsSkeleton } from './AwaitGCPDetails'

const mockStack = {
  versions: [
    {
      contents: JSON.stringify({ tfvars: 'install_id = "install-1"' }),
    },
  ],
} as any

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetails stack={mockStack} step={mockStep} installId="install-1" />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetailsSkeleton />
  </div>
)
