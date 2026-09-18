export default {
  title: 'Workflows/StepDetails/AwaitStackDetails',
}

import { AwaitStackDetails } from './AwaitStackDetails'

const appliedVersion = {
  id: 'ist-applied',
  composite_status: {
    status: 'active',
    status_human_description: 'Stack is running',
  },
  runs: [
    {
      id: 'isr-applied',
      updated_at: new Date().toISOString(),
      data_contents: { vpc_id: 'vpc-123', region: 'us-east-1' },
    },
  ],
}

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
  step_target_id: appliedVersion.id,
} as any

const mockStack = {
  versions: [appliedVersion],
  install_stack_outputs: {
    data_contents: { vpc_id: 'vpc-123', region: 'us-east-1' },
  },
} as any

// A reprovision creates a new stack version while the previous one's outputs
// are still on the stack record — the new version has no run yet.
const reprovisioningStep = {
  id: 'step-2',
  status: { status: 'active' },
  step_target_id: 'ist-new',
} as any

const reprovisioningStack = {
  versions: [
    {
      id: 'ist-new',
      composite_status: {
        status: 'generating',
        status_human_description: 'Waiting for the stack to be applied',
      },
      runs: [],
    },
    appliedVersion,
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

export const Reprovisioning = () => (
  <div className="max-w-2xl p-4">
    <AwaitStackDetails
      stack={reprovisioningStack}
      step={reprovisioningStep}
      runnerType="aws"
    />
  </div>
)

// A step old enough that its version has dropped off the stack response must
// not borrow outputs from a version that is still there.
export const VersionNotAvailable = () => (
  <div className="max-w-2xl p-4">
    <AwaitStackDetails
      stack={mockStack}
      step={{ ...mockStep, step_target_id: 'ist-aged-out' }}
      runnerType="aws"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitStackDetails
      stack={mockStack}
      step={mockStep}
      runnerType="aws"
      loading
    />
  </div>
)
