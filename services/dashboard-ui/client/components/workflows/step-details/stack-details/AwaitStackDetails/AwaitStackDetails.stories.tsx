export default {
  title: 'Workflows/StepDetails/AwaitStackDetails',
}

import { AwaitStackDetails, AwaitStackDetailsSkeleton } from './AwaitStackDetails'

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

const mockStack = {
  versions: [
    {
      composite_status: {
        status: 'active',
        status_human_description: 'Stack is running',
      },
      runs: [{ updated_at: new Date().toISOString() }],
    },
  ],
  install_stack_outputs: {
    data_contents: { vpc_id: 'vpc-123', region: 'us-east-1' },
  },
} as any

export const AWS = () => (
  <div className="max-w-2xl p-4">
    <AwaitStackDetails stack={mockStack} step={mockStep} runnerType="aws" />
  </div>
)

export const GCP = () => (
  <div className="max-w-2xl p-4">
    <AwaitStackDetails stack={mockStack} step={mockStep} runnerType="gcp" />
  </div>
)

export const Azure = () => (
  <div className="max-w-2xl p-4">
    <AwaitStackDetails stack={mockStack} step={mockStep} runnerType="azure" />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitStackDetailsSkeleton runnerType="aws" />
  </div>
)
